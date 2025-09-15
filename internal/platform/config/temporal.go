package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// TemporalConfig represents system-wide Temporal configuration
type TemporalConfig struct {
	// Connection settings
	HostPort  string `yaml:"host_port" mapstructure:"host_port"`
	Namespace string `yaml:"namespace" mapstructure:"namespace"`

	// TLS configuration
	TLS TemporalTLSConfig `yaml:"tls" mapstructure:"tls"`

	// Worker configuration
	Workers TemporalWorkersConfig `yaml:"workers" mapstructure:"workers"`

	// Client configuration
	Client TemporalClientConfig `yaml:"client" mapstructure:"client"`

	// Metrics and observability
	Metrics TemporalMetricsConfig `yaml:"metrics" mapstructure:"metrics"`

	// Feature flags
	Features TemporalFeatureConfig `yaml:"features" mapstructure:"features"`

	// Module-specific configurations
	Modules TemporalModulesConfig `yaml:"modules" mapstructure:"modules"`
}

// TemporalTLSConfig represents TLS configuration for Temporal
type TemporalTLSConfig struct {
	Enabled            bool   `yaml:"enabled" mapstructure:"enabled"`
	CertPath           string `yaml:"cert_path" mapstructure:"cert_path"`
	KeyPath            string `yaml:"key_path" mapstructure:"key_path"`
	CaPath             string `yaml:"ca_path" mapstructure:"ca_path"`
	ServerName         string `yaml:"server_name" mapstructure:"server_name"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify" mapstructure:"insecure_skip_verify"`
}

// TemporalWorkersConfig represents global worker configuration
type TemporalWorkersConfig struct {
	// Global worker settings (defaults for all modules)
	MaxConcurrentActivities      int           `yaml:"max_concurrent_activities" mapstructure:"max_concurrent_activities"`
	MaxConcurrentWorkflows       int           `yaml:"max_concurrent_workflows" mapstructure:"max_concurrent_workflows"`
	MaxConcurrentLocalActivities int           `yaml:"max_concurrent_local_activities" mapstructure:"max_concurrent_local_activities"`
	WorkerStopTimeout            time.Duration `yaml:"worker_stop_timeout" mapstructure:"worker_stop_timeout"`
	EnableLoggingInReplay        bool          `yaml:"enable_logging_in_replay" mapstructure:"enable_logging_in_replay"`
	StickyScheduleToStartTimeout time.Duration `yaml:"sticky_schedule_to_start_timeout" mapstructure:"sticky_schedule_to_start_timeout"`

	// System-wide task queues for cross-module operations
	System        TemporalTaskQueueConfig `yaml:"system" mapstructure:"system"`
	Notifications TemporalTaskQueueConfig `yaml:"notifications" mapstructure:"notifications"`
	Analytics     TemporalTaskQueueConfig `yaml:"analytics" mapstructure:"analytics"`
}

// TemporalModulesConfig represents module-specific Temporal configurations
type TemporalModulesConfig struct {
	Finance     TemporalModuleConfig `yaml:"finance" mapstructure:"finance"`
	IAM         TemporalModuleConfig `yaml:"iam" mapstructure:"iam"`
	FeatureFlag TemporalModuleConfig `yaml:"feature_flag" mapstructure:"feature_flag"`
	Audit       TemporalModuleConfig `yaml:"audit" mapstructure:"audit"`
	Tenant      TemporalModuleConfig `yaml:"tenant" mapstructure:"tenant"`
	Entity      TemporalModuleConfig `yaml:"entity" mapstructure:"entity"`
	ABAC        TemporalModuleConfig `yaml:"abac" mapstructure:"abac"`
}

// TemporalModuleConfig represents configuration for a specific module
type TemporalModuleConfig struct {
	Enabled    bool                               `yaml:"enabled" mapstructure:"enabled"`
	TaskQueues map[string]TemporalTaskQueueConfig `yaml:"task_queues" mapstructure:"task_queues"`
}

// TemporalTaskQueueConfig represents configuration for a specific task queue
type TemporalTaskQueueConfig struct {
	Enabled                      bool `yaml:"enabled" mapstructure:"enabled"`
	MaxConcurrentActivities      int  `yaml:"max_concurrent_activities" mapstructure:"max_concurrent_activities"`
	MaxConcurrentWorkflows       int  `yaml:"max_concurrent_workflows" mapstructure:"max_concurrent_workflows"`
	MaxConcurrentLocalActivities int  `yaml:"max_concurrent_local_activities" mapstructure:"max_concurrent_local_activities"`
}

// TemporalClientConfig represents client-specific configuration
type TemporalClientConfig struct {
	Identity                     string        `yaml:"identity" mapstructure:"identity"`
	DataConverter                string        `yaml:"data_converter" mapstructure:"data_converter"`
	FailureConverter             string        `yaml:"failure_converter" mapstructure:"failure_converter"`
	ContextPropagators           []string      `yaml:"context_propagators" mapstructure:"context_propagators"`
	ConnectionTimeout            time.Duration `yaml:"connection_timeout" mapstructure:"connection_timeout"`
	KeepAliveTime                time.Duration `yaml:"keep_alive_time" mapstructure:"keep_alive_time"`
	KeepAliveTimeout             time.Duration `yaml:"keep_alive_timeout" mapstructure:"keep_alive_timeout"`
	KeepAlivePermitWithoutStream bool          `yaml:"keep_alive_permit_without_stream" mapstructure:"keep_alive_permit_without_stream"`
}

// TemporalMetricsConfig represents metrics and observability configuration
type TemporalMetricsConfig struct {
	Enabled           bool          `yaml:"enabled" mapstructure:"enabled"`
	PrometheusScope   string        `yaml:"prometheus_scope" mapstructure:"prometheus_scope"`
	Tags              []string      `yaml:"tags" mapstructure:"tags"`
	ReportingInterval time.Duration `yaml:"reporting_interval" mapstructure:"reporting_interval"`
}

// TemporalFeatureConfig represents system-wide feature flags for Temporal
type TemporalFeatureConfig struct {
	EnableWorkflowShadowing    bool `yaml:"enable_workflow_shadowing" mapstructure:"enable_workflow_shadowing"`
	EnableSessionWorker        bool `yaml:"enable_session_worker" mapstructure:"enable_session_worker"`
	EnableBatchOperations      bool `yaml:"enable_batch_operations" mapstructure:"enable_batch_operations"`
	EnableMultiTenantIsolation bool `yaml:"enable_multi_tenant_isolation" mapstructure:"enable_multi_tenant_isolation"`
}

// SetTemporalDefaults sets default values for system-wide Temporal configuration
func SetTemporalDefaults(v *viper.Viper) {
	// Connection defaults
	v.SetDefault("temporal.host_port", "localhost:7233")
	v.SetDefault("temporal.namespace", "default")

	// TLS defaults
	v.SetDefault("temporal.tls.enabled", false)
	v.SetDefault("temporal.tls.cert_path", "")
	v.SetDefault("temporal.tls.key_path", "")
	v.SetDefault("temporal.tls.ca_path", "")
	v.SetDefault("temporal.tls.server_name", "")
	v.SetDefault("temporal.tls.insecure_skip_verify", false)

	// Global worker defaults
	v.SetDefault("temporal.workers.max_concurrent_activities", 100)
	v.SetDefault("temporal.workers.max_concurrent_workflows", 50)
	v.SetDefault("temporal.workers.max_concurrent_local_activities", 100)
	v.SetDefault("temporal.workers.worker_stop_timeout", 30*time.Second)
	v.SetDefault("temporal.workers.enable_logging_in_replay", true)
	v.SetDefault("temporal.workers.sticky_schedule_to_start_timeout", 5*time.Second)

	// System task queues
	v.SetDefault("temporal.workers.system.enabled", true)
	v.SetDefault("temporal.workers.system.max_concurrent_activities", 50)
	v.SetDefault("temporal.workers.system.max_concurrent_workflows", 25)
	v.SetDefault("temporal.workers.notifications.enabled", true)
	v.SetDefault("temporal.workers.notifications.max_concurrent_activities", 200)
	v.SetDefault("temporal.workers.notifications.max_concurrent_workflows", 100)
	v.SetDefault("temporal.workers.analytics.enabled", true)
	v.SetDefault("temporal.workers.analytics.max_concurrent_activities", 50)
	v.SetDefault("temporal.workers.analytics.max_concurrent_workflows", 25)

	// Module defaults - Finance
	v.SetDefault("temporal.modules.finance.enabled", true)
	v.SetDefault("temporal.modules.finance.task_queues.standard.enabled", true)
	v.SetDefault("temporal.modules.finance.task_queues.standard.max_concurrent_activities", 100)
	v.SetDefault("temporal.modules.finance.task_queues.standard.max_concurrent_workflows", 50)
	v.SetDefault("temporal.modules.finance.task_queues.high_priority.enabled", true)
	v.SetDefault("temporal.modules.finance.task_queues.high_priority.max_concurrent_activities", 50)
	v.SetDefault("temporal.modules.finance.task_queues.high_priority.max_concurrent_workflows", 25)
	v.SetDefault("temporal.modules.finance.task_queues.bulk.enabled", true)
	v.SetDefault("temporal.modules.finance.task_queues.bulk.max_concurrent_activities", 200)
	v.SetDefault("temporal.modules.finance.task_queues.bulk.max_concurrent_workflows", 10)
	v.SetDefault("temporal.modules.finance.task_queues.long_running.enabled", true)
	v.SetDefault("temporal.modules.finance.task_queues.long_running.max_concurrent_activities", 10)
	v.SetDefault("temporal.modules.finance.task_queues.long_running.max_concurrent_workflows", 5)

	// Module defaults - IAM
	v.SetDefault("temporal.modules.iam.enabled", true)
	v.SetDefault("temporal.modules.iam.task_queues.standard.enabled", true)
	v.SetDefault("temporal.modules.iam.task_queues.standard.max_concurrent_activities", 50)
	v.SetDefault("temporal.modules.iam.task_queues.standard.max_concurrent_workflows", 25)
	v.SetDefault("temporal.modules.iam.task_queues.authentication.enabled", true)
	v.SetDefault("temporal.modules.iam.task_queues.authentication.max_concurrent_activities", 200)
	v.SetDefault("temporal.modules.iam.task_queues.authentication.max_concurrent_workflows", 100)

	// Module defaults - Feature Flags
	v.SetDefault("temporal.modules.feature_flag.enabled", true)
	v.SetDefault("temporal.modules.feature_flag.task_queues.standard.enabled", true)
	v.SetDefault("temporal.modules.feature_flag.task_queues.standard.max_concurrent_activities", 100)
	v.SetDefault("temporal.modules.feature_flag.task_queues.standard.max_concurrent_workflows", 50)
	v.SetDefault("temporal.modules.feature_flag.task_queues.evaluation.enabled", true)
	v.SetDefault("temporal.modules.feature_flag.task_queues.evaluation.max_concurrent_activities", 500)
	v.SetDefault("temporal.modules.feature_flag.task_queues.evaluation.max_concurrent_workflows", 250)

	// Module defaults - Audit
	v.SetDefault("temporal.modules.audit.enabled", true)
	v.SetDefault("temporal.modules.audit.task_queues.standard.enabled", true)
	v.SetDefault("temporal.modules.audit.task_queues.standard.max_concurrent_activities", 200)
	v.SetDefault("temporal.modules.audit.task_queues.standard.max_concurrent_workflows", 100)

	// Module defaults - ABAC
	v.SetDefault("temporal.modules.abac.enabled", true)
	v.SetDefault("temporal.modules.abac.task_queues.standard.enabled", true)
	v.SetDefault("temporal.modules.abac.task_queues.standard.max_concurrent_activities", 100)
	v.SetDefault("temporal.modules.abac.task_queues.standard.max_concurrent_workflows", 50)
	v.SetDefault("temporal.modules.abac.task_queues.policy_evaluation.enabled", true)
	v.SetDefault("temporal.modules.abac.task_queues.policy_evaluation.max_concurrent_activities", 300)
	v.SetDefault("temporal.modules.abac.task_queues.policy_evaluation.max_concurrent_workflows", 150)

	// Client defaults
	v.SetDefault("temporal.client.identity", "awo-erp")
	v.SetDefault("temporal.client.data_converter", "json")
	v.SetDefault("temporal.client.failure_converter", "default")
	v.SetDefault("temporal.client.context_propagators", []string{})
	v.SetDefault("temporal.client.connection_timeout", 10*time.Second)
	v.SetDefault("temporal.client.keep_alive_time", 30*time.Second)
	v.SetDefault("temporal.client.keep_alive_timeout", 5*time.Second)
	v.SetDefault("temporal.client.keep_alive_permit_without_stream", true)

	// Metrics defaults
	v.SetDefault("temporal.metrics.enabled", true)
	v.SetDefault("temporal.metrics.prometheus_scope", "temporal_awo_erp")
	v.SetDefault("temporal.metrics.tags", []string{"service:awo-erp"})
	v.SetDefault("temporal.metrics.reporting_interval", 10*time.Second)

	// Feature defaults
	v.SetDefault("temporal.features.enable_workflow_shadowing", false)
	v.SetDefault("temporal.features.enable_session_worker", false)
	v.SetDefault("temporal.features.enable_batch_operations", true)
	v.SetDefault("temporal.features.enable_multi_tenant_isolation", true)
}

// BindTemporalEnvVars binds environment variables for Temporal configuration
func BindTemporalEnvVars(v *viper.Viper) {
	// Connection
	v.BindEnv("temporal.host_port", "TEMPORAL_HOST_PORT")
	v.BindEnv("temporal.namespace", "TEMPORAL_NAMESPACE")

	// TLS
	v.BindEnv("temporal.tls.enabled", "TEMPORAL_TLS_ENABLED")
	v.BindEnv("temporal.tls.cert_path", "TEMPORAL_TLS_CERT_PATH")
	v.BindEnv("temporal.tls.key_path", "TEMPORAL_TLS_KEY_PATH")
	v.BindEnv("temporal.tls.ca_path", "TEMPORAL_TLS_CA_PATH")
	v.BindEnv("temporal.tls.server_name", "TEMPORAL_TLS_SERVER_NAME")
	v.BindEnv("temporal.tls.insecure_skip_verify", "TEMPORAL_TLS_INSECURE_SKIP_VERIFY")

	// Workers
	v.BindEnv("temporal.workers.max_concurrent_activities", "TEMPORAL_MAX_CONCURRENT_ACTIVITIES")
	v.BindEnv("temporal.workers.max_concurrent_workflows", "TEMPORAL_MAX_CONCURRENT_WORKFLOWS")
	v.BindEnv("temporal.workers.max_concurrent_local_activities", "TEMPORAL_MAX_CONCURRENT_LOCAL_ACTIVITIES")
	v.BindEnv("temporal.workers.worker_stop_timeout", "TEMPORAL_WORKER_STOP_TIMEOUT")
	v.BindEnv("temporal.workers.enable_logging_in_replay", "TEMPORAL_ENABLE_LOGGING_IN_REPLAY")

	// System queues
	v.BindEnv("temporal.workers.system.enabled", "TEMPORAL_SYSTEM_WORKER_ENABLED")
	v.BindEnv("temporal.workers.notifications.enabled", "TEMPORAL_NOTIFICATIONS_WORKER_ENABLED")
	v.BindEnv("temporal.workers.analytics.enabled", "TEMPORAL_ANALYTICS_WORKER_ENABLED")

	// Module toggles
	v.BindEnv("temporal.modules.finance.enabled", "TEMPORAL_FINANCE_MODULE_ENABLED")
	v.BindEnv("temporal.modules.iam.enabled", "TEMPORAL_IAM_MODULE_ENABLED")
	v.BindEnv("temporal.modules.feature_flag.enabled", "TEMPORAL_FEATURE_FLAG_MODULE_ENABLED")
	v.BindEnv("temporal.modules.audit.enabled", "TEMPORAL_AUDIT_MODULE_ENABLED")
	v.BindEnv("temporal.modules.abac.enabled", "TEMPORAL_ABAC_MODULE_ENABLED")

	// Client
	v.BindEnv("temporal.client.identity", "TEMPORAL_CLIENT_IDENTITY")
	v.BindEnv("temporal.client.connection_timeout", "TEMPORAL_CONNECTION_TIMEOUT")
	v.BindEnv("temporal.client.keep_alive_time", "TEMPORAL_KEEP_ALIVE_TIME")
	v.BindEnv("temporal.client.keep_alive_timeout", "TEMPORAL_KEEP_ALIVE_TIMEOUT")

	// Metrics
	v.BindEnv("temporal.metrics.enabled", "TEMPORAL_METRICS_ENABLED")
	v.BindEnv("temporal.metrics.prometheus_scope", "TEMPORAL_PROMETHEUS_SCOPE")
	v.BindEnv("temporal.metrics.reporting_interval", "TEMPORAL_METRICS_REPORTING_INTERVAL")

	// Features
	v.BindEnv("temporal.features.enable_workflow_shadowing", "TEMPORAL_ENABLE_WORKFLOW_SHADOWING")
	v.BindEnv("temporal.features.enable_session_worker", "TEMPORAL_ENABLE_SESSION_WORKER")
	v.BindEnv("temporal.features.enable_batch_operations", "TEMPORAL_ENABLE_BATCH_OPERATIONS")
	v.BindEnv("temporal.features.enable_multi_tenant_isolation", "TEMPORAL_ENABLE_MULTI_TENANT_ISOLATION")
}

// Validate validates the entire Temporal configuration
func (t *TemporalConfig) Validate() error {
	if t.HostPort == "" {
		return fmt.Errorf("temporal host_port cannot be empty")
	}

	if t.Namespace == "" {
		return fmt.Errorf("temporal namespace cannot be empty")
	}

	// Validate worker configuration
	if err := t.Workers.Validate(); err != nil {
		return fmt.Errorf("temporal workers config validation failed: %w", err)
	}

	// Validate client configuration
	if err := t.Client.Validate(); err != nil {
		return fmt.Errorf("temporal client config validation failed: %w", err)
	}

	// Validate TLS configuration
	if err := t.TLS.Validate(); err != nil {
		return fmt.Errorf("temporal TLS config validation failed: %w", err)
	}

	// Validate modules configuration
	if err := t.Modules.Validate(); err != nil {
		return fmt.Errorf("temporal modules config validation failed: %w", err)
	}

	return nil
}

// Validate validates worker configuration
func (w *TemporalWorkersConfig) Validate() error {
	if w.MaxConcurrentActivities <= 0 {
		return fmt.Errorf("max_concurrent_activities must be greater than 0")
	}

	if w.MaxConcurrentWorkflows <= 0 {
		return fmt.Errorf("max_concurrent_workflows must be greater than 0")
	}

	if w.WorkerStopTimeout <= 0 {
		return fmt.Errorf("worker_stop_timeout must be greater than 0")
	}

	// Validate system task queues
	if err := w.System.Validate("system"); err != nil {
		return err
	}
	if err := w.Notifications.Validate("notifications"); err != nil {
		return err
	}
	if err := w.Analytics.Validate("analytics"); err != nil {
		return err
	}

	return nil
}

// Validate validates modules configuration
func (m *TemporalModulesConfig) Validate() error {
	modules := map[string]TemporalModuleConfig{
		"finance":      m.Finance,
		"iam":          m.IAM,
		"feature_flag": m.FeatureFlag,
		"audit":        m.Audit,
		"tenant":       m.Tenant,
		"entity":       m.Entity,
		"abac":         m.ABAC,
	}

	for moduleName, moduleConfig := range modules {
		if err := moduleConfig.Validate(moduleName); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates module configuration
func (mc *TemporalModuleConfig) Validate(moduleName string) error {
	if mc.Enabled {
		for queueName, queueConfig := range mc.TaskQueues {
			if err := queueConfig.Validate(fmt.Sprintf("%s.%s", moduleName, queueName)); err != nil {
				return err
			}
		}
	}
	return nil
}

// Validate validates task queue configuration
func (tq *TemporalTaskQueueConfig) Validate(queueName string) error {
	if tq.Enabled {
		if tq.MaxConcurrentActivities <= 0 {
			return fmt.Errorf("task queue %s: max_concurrent_activities must be greater than 0", queueName)
		}
		if tq.MaxConcurrentWorkflows <= 0 {
			return fmt.Errorf("task queue %s: max_concurrent_workflows must be greater than 0", queueName)
		}
	}
	return nil
}

// Validate validates client configuration
func (c *TemporalClientConfig) Validate() error {
	if c.Identity == "" {
		return fmt.Errorf("client identity cannot be empty")
	}

	if c.ConnectionTimeout <= 0 {
		return fmt.Errorf("connection_timeout must be greater than 0")
	}

	return nil
}

// Validate validates TLS configuration
func (t *TemporalTLSConfig) Validate() error {
	if t.Enabled {
		if t.CertPath == "" {
			return fmt.Errorf("TLS cert_path cannot be empty when TLS is enabled")
		}
		if t.KeyPath == "" {
			return fmt.Errorf("TLS key_path cannot be empty when TLS is enabled")
		}
	}
	return nil
}

// Helper methods for getting module configurations

// GetEnabledModules returns list of enabled modules
func (t *TemporalConfig) GetEnabledModules() []string {
	var enabled []string

	if t.Modules.Finance.Enabled {
		enabled = append(enabled, "finance")
	}
	if t.Modules.IAM.Enabled {
		enabled = append(enabled, "iam")
	}
	if t.Modules.FeatureFlag.Enabled {
		enabled = append(enabled, "feature_flag")
	}
	if t.Modules.Audit.Enabled {
		enabled = append(enabled, "audit")
	}
	if t.Modules.Tenant.Enabled {
		enabled = append(enabled, "tenant")
	}
	if t.Modules.Entity.Enabled {
		enabled = append(enabled, "entity")
	}
	if t.Modules.ABAC.Enabled {
		enabled = append(enabled, "abac")
	}

	return enabled
}

// GetModuleConfig returns configuration for a specific module
func (t *TemporalConfig) GetModuleConfig(moduleName string) *TemporalModuleConfig {
	switch moduleName {
	case "finance":
		return &t.Modules.Finance
	case "iam":
		return &t.Modules.IAM
	case "feature_flag":
		return &t.Modules.FeatureFlag
	case "audit":
		return &t.Modules.Audit
	case "tenant":
		return &t.Modules.Tenant
	case "entity":
		return &t.Modules.Entity
	case "abac":
		return &t.Modules.ABAC
	default:
		return nil
	}
}

// IsModuleEnabled checks if a specific module is enabled
func (t *TemporalConfig) IsModuleEnabled(moduleName string) bool {
	if config := t.GetModuleConfig(moduleName); config != nil {
		return config.Enabled
	}
	return false
}
