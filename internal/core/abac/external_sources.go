package abac

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	ldapv3 "github.com/go-ldap/ldap/v3"
	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// ExternalAttributeSourceManager manages external attribute sources
type ExternalAttributeSourceManager interface {
	// Source Registration
	RegisterLDAPSource(ctx context.Context, req *RegisterLDAPSourceRequest) (*ExternalAttributeSource, error)
	RegisterRESTAPISource(ctx context.Context, req *RegisterRESTAPISourceRequest) (*ExternalAttributeSource, error)
	RegisterDatabaseSource(ctx context.Context, req *RegisterDatabaseSourceRequest) (*ExternalAttributeSource, error)
	RegisterCustomSource(ctx context.Context, req *RegisterCustomSourceRequest) (*ExternalAttributeSource, error)

	// Source Management
	UpdateAttributeSource(ctx context.Context, req *UpdateAttributeSourceRequest) (*ExternalAttributeSource, error)
	DeleteAttributeSource(ctx context.Context, sourceID uuid.UUID) error
	GetAttributeSource(ctx context.Context, sourceID uuid.UUID) (*ExternalAttributeSourceDetails, error)
	ListAttributeSources(ctx context.Context, req *ListAttributeSourcesRequest) (*AttributeSourceListResult, error)

	// Connectivity Testing
	TestSourceConnection(ctx context.Context, sourceID uuid.UUID) (*SourceConnectionTestResult, error)
	ValidateSourceConfiguration(ctx context.Context, req *ValidateSourceConfigRequest) (*SourceConfigValidationResult, error)

	// Attribute Retrieval
	FetchAttributes(ctx context.Context, req *FetchAttributesRequest) (*ExternalAttributesResult, error)
	BatchFetchAttributes(ctx context.Context, req *BatchFetchAttributesRequest) (*BatchExternalAttributesResult, error)

	// Source Monitoring
	GetSourceHealth(ctx context.Context, sourceID uuid.UUID) (*ExternalSourceHealthStatus, error)
	GetSourceMetrics(ctx context.Context, req *SourceMetricsRequest) (*ExternalSourceMetrics, error)

	// Synchronization
	SynchronizeAttributes(ctx context.Context, req *SynchronizeAttributesRequest) (*SynchronizationResult, error)
	ScheduleAttributeSync(ctx context.Context, req *ScheduleSyncRequest) (*SyncSchedule, error)
}

// externalAttributeSourceManager implements ExternalAttributeSourceManager
type externalAttributeSourceManager struct {
	sourceRepo        repository.AttributeSourceRepository
	connectorRegistry *SourceConnectorRegistry
	healthMonitor     *ExternalSourceHealthMonitor
	syncCoordinator   *AttributeSyncCoordinator

	// Connection pools
	ldapConnections map[uuid.UUID]*LDAPConnectionPool
	httpClients     map[uuid.UUID]*http.Client

	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
	mutex   sync.RWMutex
}

// NewExternalAttributeSourceManager creates a new external attribute source manager
func NewExternalAttributeSourceManager(
	sourceRepo repository.AttributeSourceRepository,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) ExternalAttributeSourceManager {
	return &externalAttributeSourceManager{
		sourceRepo:        sourceRepo,
		connectorRegistry: NewSourceConnectorRegistry(),
		healthMonitor:     NewExternalSourceHealthMonitor(),
		syncCoordinator:   NewAttributeSyncCoordinator(),
		ldapConnections:   make(map[uuid.UUID]*LDAPConnectionPool),
		httpClients:       make(map[uuid.UUID]*http.Client),
		logger:            logger,
		metrics:           metrics,
		tracer:            tracer,
	}
}

// External Source Types

type ExternalSourceType string

const (
	ExternalSourceTypeLDAP     ExternalSourceType = "ldap"
	ExternalSourceTypeRESTAPI  ExternalSourceType = "rest_api"
	ExternalSourceTypeDatabase ExternalSourceType = "database"
	ExternalSourceTypeCustom   ExternalSourceType = "custom"
)

type ExternalAttributeSource struct {
	ID               uuid.UUID                      `json:"id"`
	Name             string                         `json:"name"`
	Description      *string                        `json:"description,omitempty"`
	SourceType       ExternalSourceType             `json:"source_type"`
	Configuration    ExternalSourceConfiguration    `json:"configuration"`
	IsActive         bool                           `json:"is_active"`
	Priority         int                            `json:"priority"`
	Timeout          time.Duration                  `json:"timeout"`
	RetryPolicy      RetryPolicy                    `json:"retry_policy"`
	CacheSettings    ExternalSourceCacheSettings    `json:"cache_settings"`
	SecuritySettings ExternalSourceSecuritySettings `json:"security_settings"`
	SyncSettings     SynchronizationSettings        `json:"sync_settings"`
	CreatedAt        time.Time                      `json:"created_at"`
	UpdatedAt        time.Time                      `json:"updated_at"`
	CreatedBy        *uuid.UUID                     `json:"created_by,omitempty"`
}

type ExternalSourceConfiguration struct {
	// LDAP Configuration
	LDAPConfig *LDAPConfiguration `json:"ldap_config,omitempty"`

	// REST API Configuration
	RESTAPIConfig *RESTAPIConfiguration `json:"rest_api_config,omitempty"`

	// Database Configuration
	DatabaseConfig *DatabaseConfiguration `json:"database_config,omitempty"`

	// Custom Source Configuration
	CustomConfig map[string]interface{} `json:"custom_config,omitempty"`
}

type LDAPConfiguration struct {
	Host              string            `json:"host" validate:"required"`
	Port              int               `json:"port" validate:"required,min=1,max=65535"`
	BaseDN            string            `json:"base_dn" validate:"required"`
	BindDN            string            `json:"bind_dn" validate:"required"`
	BindPassword      string            `json:"bind_password" validate:"required"`
	SearchFilter      string            `json:"search_filter" validate:"required"`
	AttributeMappings map[string]string `json:"attribute_mappings" validate:"required"`
	UseTLS            bool              `json:"use_tls"`
	TLSConfig         *TLSConfiguration `json:"tls_config,omitempty"`
	ConnectionPool    PoolConfiguration `json:"connection_pool"`
}

type RESTAPIConfiguration struct {
	BaseURL        string                 `json:"base_url" validate:"required,url"`
	AuthMethod     APIAuthMethod          `json:"auth_method" validate:"required"`
	AuthConfig     APIAuthConfiguration   `json:"auth_config"`
	Endpoints      map[string]APIEndpoint `json:"endpoints" validate:"required"`
	DefaultHeaders map[string]string      `json:"default_headers,omitempty"`
	Timeout        time.Duration          `json:"timeout"`
	RateLimiting   *RateLimitConfig       `json:"rate_limiting,omitempty"`
}

type DatabaseConfiguration struct {
	Driver           string            `json:"driver" validate:"required"`
	ConnectionString string            `json:"connection_string" validate:"required"`
	QueryTemplates   map[string]string `json:"query_templates" validate:"required"`
	ColumnMappings   map[string]string `json:"column_mappings" validate:"required"`
	ConnectionPool   PoolConfiguration `json:"connection_pool"`
}

// LDAP Source Implementation

type LDAPAttributeConnector struct {
	config   *LDAPConfiguration
	connPool *LDAPConnectionPool
	logger   logger.Logger
	metrics  metrics.MetricsProvider
}

func NewLDAPAttributeConnector(config *LDAPConfiguration, logger logger.Logger, metrics metrics.MetricsProvider) *LDAPAttributeConnector {
	return &LDAPAttributeConnector{
		config:   config,
		connPool: NewLDAPConnectionPool(config),
		logger:   logger,
		metrics:  metrics,
	}
}

func (lac *LDAPAttributeConnector) FetchUserAttributes(ctx context.Context, userID string) (map[string]interface{}, error) {
	startTime := time.Now()
	defer func() {
		lac.metrics.RecordHistogram("abac.external_source.ldap.fetch_duration",
			time.Since(startTime).Seconds(), map[string]string{"operation": "fetch_user_attributes"})
	}()

	conn, err := lac.connPool.GetConnection(ctx)
	if err != nil {
		lac.logger.Error("Failed to get LDAP connection", "error", err)
		return nil, errors.Wrap(err, "failed to get LDAP connection")
	}
	defer lac.connPool.ReturnConnection(conn)

	// Build search filter
	searchFilter := fmt.Sprintf(lac.config.SearchFilter, userID)

	// Define attributes to retrieve
	var attributes []string
	for ldapAttr := range lac.config.AttributeMappings {
		attributes = append(attributes, ldapAttr)
	}

	searchRequest := ldapv3.NewSearchRequest(
		lac.config.BaseDN,
		ldapv3.ScopeWholeSubtree,
		ldapv3.NeverDerefAliases,
		0, 0, false,
		searchFilter,
		attributes,
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		lac.logger.Error("LDAP search failed", "error", err, "filter", searchFilter)
		return nil, errors.Wrap(err, "LDAP search failed")
	}

	if len(result.Entries) == 0 {
		return nil, errors.NewNotFoundError("user not found in LDAP", "user_id", userID)
	}

	// Map LDAP attributes to internal attributes
	attributes_map := make(map[string]interface{})
	entry := result.Entries[0]

	for ldapAttr, internalAttr := range lac.config.AttributeMappings {
		values := entry.GetAttributeValues(ldapAttr)
		if len(values) > 0 {
			if len(values) == 1 {
				attributes_map[internalAttr] = values[0]
			} else {
				attributes_map[internalAttr] = values
			}
		}
	}

	lac.logger.Debug("Successfully fetched LDAP attributes",
		"user_id", userID, "attribute_count", len(attributes_map))

	return attributes_map, nil
}

// REST API Source Implementation

type RESTAPIAttributeConnector struct {
	config     *RESTAPIConfiguration
	httpClient *http.Client
	logger     logger.Logger
	metrics    metrics.MetricsProvider
}

func NewRESTAPIAttributeConnector(config *RESTAPIConfiguration, logger logger.Logger, metrics metrics.MetricsProvider) *RESTAPIAttributeConnector {
	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &RESTAPIAttributeConnector{
		config:     config,
		httpClient: client,
		logger:     logger,
		metrics:    metrics,
	}
}

func (rac *RESTAPIAttributeConnector) FetchUserAttributes(ctx context.Context, userID string) (map[string]interface{}, error) {
	startTime := time.Now()
	defer func() {
		rac.metrics.RecordHistogram("abac.external_source.rest_api.fetch_duration",
			time.Since(startTime).Seconds(), map[string]string{"operation": "fetch_user_attributes"})
	}()

	endpoint, exists := rac.config.Endpoints["user_attributes"]
	if !exists {
		return nil, errors.NewInvalidInputError("user_attributes endpoint not configured", "endpoint", "user_attributes")
	}

	// Build request URL
	url := fmt.Sprintf("%s%s", rac.config.BaseURL, fmt.Sprintf(endpoint.Path, userID))

	req, err := http.NewRequestWithContext(ctx, endpoint.Method, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create HTTP request")
	}

	// Add authentication
	if err := rac.addAuthentication(req); err != nil {
		return nil, errors.Wrap(err, "failed to add authentication")
	}

	// Add default headers
	for key, value := range rac.config.DefaultHeaders {
		req.Header.Set(key, value)
	}

	resp, err := rac.httpClient.Do(req)
	if err != nil {
		rac.logger.Error("HTTP request failed", "error", err, "url", url)
		return nil, errors.Wrap(err, "HTTP request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rac.logger.Error("HTTP request returned non-OK status",
			"status_code", resp.StatusCode, "url", url)
		return nil, errors.NewExternalServiceError("API request failed",
			"status_code", resp.StatusCode, "url", url)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to decode JSON response")
	}

	// Transform response according to endpoint mapping
	attributes := make(map[string]interface{})
	if endpoint.ResponseMapping != nil {
		for apiField, internalAttr := range endpoint.ResponseMapping {
			if value, exists := response[apiField]; exists {
				attributes[internalAttr] = value
			}
		}
	} else {
		// Use response as-is if no mapping defined
		attributes = response
	}

	rac.logger.Debug("Successfully fetched REST API attributes",
		"user_id", userID, "attribute_count", len(attributes))

	return attributes, nil
}

func (rac *RESTAPIAttributeConnector) addAuthentication(req *http.Request) error {
	switch rac.config.AuthMethod {
	case APIAuthMethodAPIKey:
		if apiKeyConfig, ok := rac.config.AuthConfig.(*APIKeyAuthConfig); ok {
			switch apiKeyConfig.Location {
			case APIKeyLocationHeader:
				req.Header.Set(apiKeyConfig.KeyName, apiKeyConfig.KeyValue)
			case APIKeyLocationQuery:
				q := req.URL.Query()
				q.Set(apiKeyConfig.KeyName, apiKeyConfig.KeyValue)
				req.URL.RawQuery = q.Encode()
			}
		}
	case APIAuthMethodBearer:
		if bearerConfig, ok := rac.config.AuthConfig.(*BearerTokenAuthConfig); ok {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearerConfig.Token))
		}
	case APIAuthMethodBasic:
		if basicConfig, ok := rac.config.AuthConfig.(*BasicAuthConfig); ok {
			req.SetBasicAuth(basicConfig.Username, basicConfig.Password)
		}
	}
	return nil
}

// Source Registration Methods

func (easm *externalAttributeSourceManager) RegisterLDAPSource(ctx context.Context, req *RegisterLDAPSourceRequest) (*ExternalAttributeSource, error) {
	span := easm.tracer.StartSpan(ctx, "ExternalAttributeSourceManager.RegisterLDAPSource")
	defer span.End()

	if err := easm.validateLDAPConfiguration(req.LDAPConfig); err != nil {
		return nil, errors.Wrap(err, "invalid LDAP configuration")
	}

	source := &ExternalAttributeSource{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		SourceType:  ExternalSourceTypeLDAP,
		Configuration: ExternalSourceConfiguration{
			LDAPConfig: req.LDAPConfig,
		},
		IsActive:         req.IsActive,
		Priority:         req.Priority,
		Timeout:          req.Timeout,
		RetryPolicy:      req.RetryPolicy,
		CacheSettings:    req.CacheSettings,
		SecuritySettings: req.SecuritySettings,
		SyncSettings:     req.SyncSettings,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		CreatedBy:        req.CreatedBy,
	}

	// Create LDAP connection pool
	easm.mutex.Lock()
	easm.ldapConnections[source.ID] = NewLDAPConnectionPool(req.LDAPConfig)
	easm.mutex.Unlock()

	// Register connector
	connector := NewLDAPAttributeConnector(req.LDAPConfig, easm.logger, easm.metrics)
	easm.connectorRegistry.RegisterConnector(source.ID, connector)

	// Start health monitoring
	easm.healthMonitor.StartMonitoring(source.ID, connector)

	easm.logger.Info("LDAP attribute source registered successfully",
		"source_id", source.ID, "source_name", source.Name)

	return source, nil
}

func (easm *externalAttributeSourceManager) RegisterRESTAPISource(ctx context.Context, req *RegisterRESTAPISourceRequest) (*ExternalAttributeSource, error) {
	span := easm.tracer.StartSpan(ctx, "ExternalAttributeSourceManager.RegisterRESTAPISource")
	defer span.End()

	if err := easm.validateRESTAPIConfiguration(req.RESTAPIConfig); err != nil {
		return nil, errors.Wrap(err, "invalid REST API configuration")
	}

	source := &ExternalAttributeSource{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		SourceType:  ExternalSourceTypeRESTAPI,
		Configuration: ExternalSourceConfiguration{
			RESTAPIConfig: req.RESTAPIConfig,
		},
		IsActive:         req.IsActive,
		Priority:         req.Priority,
		Timeout:          req.Timeout,
		RetryPolicy:      req.RetryPolicy,
		CacheSettings:    req.CacheSettings,
		SecuritySettings: req.SecuritySettings,
		SyncSettings:     req.SyncSettings,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		CreatedBy:        req.CreatedBy,
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: req.RESTAPIConfig.Timeout,
	}

	easm.mutex.Lock()
	easm.httpClients[source.ID] = client
	easm.mutex.Unlock()

	// Register connector
	connector := NewRESTAPIAttributeConnector(req.RESTAPIConfig, easm.logger, easm.metrics)
	easm.connectorRegistry.RegisterConnector(source.ID, connector)

	// Start health monitoring
	easm.healthMonitor.StartMonitoring(source.ID, connector)

	easm.logger.Info("REST API attribute source registered successfully",
		"source_id", source.ID, "source_name", source.Name)

	return source, nil
}

// Supporting Types and Configurations

type RegisterLDAPSourceRequest struct {
	Name             string                         `json:"name" validate:"required,min=1,max=100"`
	Description      *string                        `json:"description,omitempty"`
	LDAPConfig       *LDAPConfiguration             `json:"ldap_config" validate:"required"`
	IsActive         bool                           `json:"is_active"`
	Priority         int                            `json:"priority"`
	Timeout          time.Duration                  `json:"timeout"`
	RetryPolicy      RetryPolicy                    `json:"retry_policy"`
	CacheSettings    ExternalSourceCacheSettings    `json:"cache_settings"`
	SecuritySettings ExternalSourceSecuritySettings `json:"security_settings"`
	SyncSettings     SynchronizationSettings        `json:"sync_settings"`
	CreatedBy        *uuid.UUID                     `json:"created_by,omitempty"`
}

type RegisterRESTAPISourceRequest struct {
	Name             string                         `json:"name" validate:"required,min=1,max=100"`
	Description      *string                        `json:"description,omitempty"`
	RESTAPIConfig    *RESTAPIConfiguration          `json:"rest_api_config" validate:"required"`
	IsActive         bool                           `json:"is_active"`
	Priority         int                            `json:"priority"`
	Timeout          time.Duration                  `json:"timeout"`
	RetryPolicy      RetryPolicy                    `json:"retry_policy"`
	CacheSettings    ExternalSourceCacheSettings    `json:"cache_settings"`
	SecuritySettings ExternalSourceSecuritySettings `json:"security_settings"`
	SyncSettings     SynchronizationSettings        `json:"sync_settings"`
	CreatedBy        *uuid.UUID                     `json:"created_by,omitempty"`
}

type RegisterDatabaseSourceRequest struct {
	Name             string                         `json:"name" validate:"required,min=1,max=100"`
	Description      *string                        `json:"description,omitempty"`
	DatabaseConfig   *DatabaseConfiguration         `json:"database_config" validate:"required"`
	IsActive         bool                           `json:"is_active"`
	Priority         int                            `json:"priority"`
	Timeout          time.Duration                  `json:"timeout"`
	RetryPolicy      RetryPolicy                    `json:"retry_policy"`
	CacheSettings    ExternalSourceCacheSettings    `json:"cache_settings"`
	SecuritySettings ExternalSourceSecuritySettings `json:"security_settings"`
	SyncSettings     SynchronizationSettings        `json:"sync_settings"`
	CreatedBy        *uuid.UUID                     `json:"created_by,omitempty"`
}

type RegisterCustomSourceRequest struct {
	Name             string                         `json:"name" validate:"required,min=1,max=100"`
	Description      *string                        `json:"description,omitempty"`
	CustomConfig     map[string]interface{}         `json:"custom_config" validate:"required"`
	ConnectorType    string                         `json:"connector_type" validate:"required"`
	IsActive         bool                           `json:"is_active"`
	Priority         int                            `json:"priority"`
	Timeout          time.Duration                  `json:"timeout"`
	RetryPolicy      RetryPolicy                    `json:"retry_policy"`
	CacheSettings    ExternalSourceCacheSettings    `json:"cache_settings"`
	SecuritySettings ExternalSourceSecuritySettings `json:"security_settings"`
	SyncSettings     SynchronizationSettings        `json:"sync_settings"`
	CreatedBy        *uuid.UUID                     `json:"created_by,omitempty"`
}

// API Authentication Types

type APIAuthMethod string

const (
	APIAuthMethodNone   APIAuthMethod = "none"
	APIAuthMethodAPIKey APIAuthMethod = "api_key"
	APIAuthMethodBearer APIAuthMethod = "bearer"
	APIAuthMethodBasic  APIAuthMethod = "basic"
	APIAuthMethodOAuth2 APIAuthMethod = "oauth2"
)

type APIAuthConfiguration interface {
	GetAuthMethod() APIAuthMethod
}

type APIKeyAuthConfig struct {
	KeyName  string         `json:"key_name" validate:"required"`
	KeyValue string         `json:"key_value" validate:"required"`
	Location APIKeyLocation `json:"location" validate:"required"`
}

func (c *APIKeyAuthConfig) GetAuthMethod() APIAuthMethod {
	return APIAuthMethodAPIKey
}

type APIKeyLocation string

const (
	APIKeyLocationHeader APIKeyLocation = "header"
	APIKeyLocationQuery  APIKeyLocation = "query"
)

type BearerTokenAuthConfig struct {
	Token string `json:"token" validate:"required"`
}

func (c *BearerTokenAuthConfig) GetAuthMethod() APIAuthMethod {
	return APIAuthMethodBearer
}

type BasicAuthConfig struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (c *BasicAuthConfig) GetAuthMethod() APIAuthMethod {
	return APIAuthMethodBasic
}

type APIEndpoint struct {
	Path            string            `json:"path" validate:"required"`
	Method          string            `json:"method" validate:"required"`
	ResponseMapping map[string]string `json:"response_mapping,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
}

// Connection Pool and Health Monitoring

type LDAPConnectionPool struct {
	config      *LDAPConfiguration
	connections chan *ldapv3.Conn
	mutex       sync.Mutex
	logger      logger.Logger
}

func NewLDAPConnectionPool(config *LDAPConfiguration) *LDAPConnectionPool {
	return &LDAPConnectionPool{
		config:      config,
		connections: make(chan *ldapv3.Conn, config.ConnectionPool.MaxSize),
	}
}

func (pool *LDAPConnectionPool) GetConnection(ctx context.Context) (*ldapv3.Conn, error) {
	select {
	case conn := <-pool.connections:
		return conn, nil
	default:
		return pool.createConnection()
	}
}

func (pool *LDAPConnectionPool) ReturnConnection(conn *ldapv3.Conn) {
	select {
	case pool.connections <- conn:
	default:
		conn.Close()
	}
}

func (pool *LDAPConnectionPool) createConnection() (*ldapv3.Conn, error) {
	address := fmt.Sprintf("%s:%d", pool.config.Host, pool.config.Port)

	var conn *ldapv3.Conn
	var err error

	if pool.config.UseTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		if pool.config.TLSConfig != nil {
			tlsConfig.InsecureSkipVerify = pool.config.TLSConfig.InsecureSkipVerify
		}
		conn, err = ldapv3.DialTLS("tcp", address, tlsConfig)
	} else {
		conn, err = ldapv3.Dial("tcp", address)
	}

	if err != nil {
		return nil, err
	}

	// Bind with credentials
	err = conn.Bind(pool.config.BindDN, pool.config.BindPassword)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

// Configuration Types

type TLSConfiguration struct {
	InsecureSkipVerify bool `json:"insecure_skip_verify"`
}

type PoolConfiguration struct {
	MaxSize     int           `json:"max_size"`
	MinSize     int           `json:"min_size"`
	MaxIdleTime time.Duration `json:"max_idle_time"`
}

type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

type ExternalSourceCacheSettings struct {
	EnableCaching   bool          `json:"enable_caching"`
	CacheTTL        time.Duration `json:"cache_ttl"`
	CacheSize       int           `json:"cache_size"`
	RefreshInterval time.Duration `json:"refresh_interval"`
}

type ExternalSourceSecuritySettings struct {
	EncryptInTransit bool     `json:"encrypt_in_transit"`
	EncryptAtRest    bool     `json:"encrypt_at_rest"`
	AllowedCIDRs     []string `json:"allowed_cidrs,omitempty"`
	RequiredClaims   []string `json:"required_claims,omitempty"`
}

type SynchronizationSettings struct {
	EnableAutoSync bool          `json:"enable_auto_sync"`
	SyncInterval   time.Duration `json:"sync_interval"`
	SyncBatchSize  int           `json:"sync_batch_size"`
	SyncStrategy   SyncStrategy  `json:"sync_strategy"`
}

type SyncStrategy string

const (
	SyncStrategyIncremental SyncStrategy = "incremental"
	SyncStrategyFull        SyncStrategy = "full"
	SyncStrategyDelta       SyncStrategy = "delta"
)

type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
}

// Validation Methods

func (easm *externalAttributeSourceManager) validateLDAPConfiguration(config *LDAPConfiguration) error {
	if config == nil {
		return errors.NewInvalidInputError("LDAP configuration is required", "config", "nil")
	}

	if config.Host == "" {
		return errors.NewInvalidInputError("LDAP host is required", "host", "")
	}

	if config.Port <= 0 || config.Port > 65535 {
		return errors.NewInvalidInputError("LDAP port must be between 1 and 65535", "port", config.Port)
	}

	if config.BaseDN == "" {
		return errors.NewInvalidInputError("LDAP base DN is required", "base_dn", "")
	}

	if config.BindDN == "" {
		return errors.NewInvalidInputError("LDAP bind DN is required", "bind_dn", "")
	}

	if config.SearchFilter == "" {
		return errors.NewInvalidInputError("LDAP search filter is required", "search_filter", "")
	}

	if len(config.AttributeMappings) == 0 {
		return errors.NewInvalidInputError("LDAP attribute mappings are required", "attribute_mappings", "empty")
	}

	return nil
}

func (easm *externalAttributeSourceManager) validateRESTAPIConfiguration(config *RESTAPIConfiguration) error {
	if config == nil {
		return errors.NewInvalidInputError("REST API configuration is required", "config", "nil")
	}

	if config.BaseURL == "" {
		return errors.NewInvalidInputError("REST API base URL is required", "base_url", "")
	}

	if len(config.Endpoints) == 0 {
		return errors.NewInvalidInputError("REST API endpoints are required", "endpoints", "empty")
	}

	return nil
}

// Helper Components

type SourceConnectorRegistry struct {
	connectors map[uuid.UUID]interface{}
	mutex      sync.RWMutex
}

func NewSourceConnectorRegistry() *SourceConnectorRegistry {
	return &SourceConnectorRegistry{
		connectors: make(map[uuid.UUID]interface{}),
	}
}

func (scr *SourceConnectorRegistry) RegisterConnector(sourceID uuid.UUID, connector interface{}) {
	scr.mutex.Lock()
	defer scr.mutex.Unlock()
	scr.connectors[sourceID] = connector
}

type ExternalSourceHealthMonitor struct {
	healthStatus map[uuid.UUID]*ExternalSourceHealthStatus
	mutex        sync.RWMutex
}

func NewExternalSourceHealthMonitor() *ExternalSourceHealthMonitor {
	return &ExternalSourceHealthMonitor{
		healthStatus: make(map[uuid.UUID]*ExternalSourceHealthStatus),
	}
}

func (eshm *ExternalSourceHealthMonitor) StartMonitoring(sourceID uuid.UUID, connector interface{}) {
	// Implementation for health monitoring
}

type AttributeSyncCoordinator struct {
	syncJobs map[uuid.UUID]*SyncJob
	mutex    sync.RWMutex
}

func NewAttributeSyncCoordinator() *AttributeSyncCoordinator {
	return &AttributeSyncCoordinator{
		syncJobs: make(map[uuid.UUID]*SyncJob),
	}
}

// Result Types

type ExternalAttributeSourceDetails struct {
	Source             *ExternalAttributeSource  `json:"source"`
	ConnectionStatus   ConnectionStatus          `json:"connection_status"`
	LastSyncTime       *time.Time                `json:"last_sync_time,omitempty"`
	AttributeCount     int                       `json:"attribute_count"`
	ErrorCount         int                       `json:"error_count"`
	PerformanceMetrics *SourcePerformanceMetrics `json:"performance_metrics,omitempty"`
}

type ConnectionStatus string

const (
	ConnectionStatusHealthy   ConnectionStatus = "healthy"
	ConnectionStatusDegraded  ConnectionStatus = "degraded"
	ConnectionStatusUnhealthy ConnectionStatus = "unhealthy"
	ConnectionStatusUnknown   ConnectionStatus = "unknown"
)

type ExternalSourceHealthStatus struct {
	SourceID         uuid.UUID        `json:"source_id"`
	Status           ConnectionStatus `json:"status"`
	LastCheckTime    time.Time        `json:"last_check_time"`
	ResponseTime     time.Duration    `json:"response_time"`
	ErrorMessage     *string          `json:"error_message,omitempty"`
	SuccessfulChecks int              `json:"successful_checks"`
	FailedChecks     int              `json:"failed_checks"`
}

type SourcePerformanceMetrics struct {
	AverageResponseTime time.Duration `json:"average_response_time"`
	SuccessRate         float64       `json:"success_rate"`
	ThroughputPerSecond float64       `json:"throughput_per_second"`
	CacheHitRate        float64       `json:"cache_hit_rate"`
}

type SyncJob struct {
	ID       uuid.UUID  `json:"id"`
	SourceID uuid.UUID  `json:"source_id"`
	Status   SyncStatus `json:"status"`
	Progress float64    `json:"progress"`
}

type SyncStatus string

const (
	SyncStatusPending   SyncStatus = "pending"
	SyncStatusRunning   SyncStatus = "running"
	SyncStatusCompleted SyncStatus = "completed"
	SyncStatusFailed    SyncStatus = "failed"
)
