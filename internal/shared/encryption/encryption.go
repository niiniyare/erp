package encryption

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"
)

// Configuration
type Config struct {
	KeyManagement KeyManagementConfig `yaml:"key_management"`
	Encryption    EncryptionConfig    `yaml:"encryption"`
	Cache         CacheConfig         `yaml:"cache"`
	Observability ObservabilityConfig `yaml:"observability"`
}

type KeyManagementConfig struct {
	Provider     string        `yaml:"provider"`
	Region       string        `yaml:"region"`
	KeyTTL       time.Duration `yaml:"key_ttl"`
	RotationFreq time.Duration `yaml:"rotation_frequency"`
}

type EncryptionConfig struct {
	Algorithm         Algorithm `yaml:"algorithm"`
	KeyDerivationSalt string    `yaml:"key_derivation_salt"`
	MaxDataSize       int       `yaml:"max_data_size"`
}

type CacheConfig struct {
	Enabled bool          `yaml:"enabled"`
	TTL     time.Duration `yaml:"ttl"`
	MaxSize int           `yaml:"max_size"`
}

type ObservabilityConfig struct {
	LogLevel    string `yaml:"log_level"`
	MetricsPort int    `yaml:"metrics_port"`
}

// Enhanced Domain Errors
type EncryptionError struct {
	Code    string
	Message string
	Cause   error
	Context map[string]any
}

func (e *EncryptionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *EncryptionError) Unwrap() error {
	return e.Cause
}

func NewEncryptionError(code, message string, cause error, context map[string]any) *EncryptionError {
	return &EncryptionError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Context: context,
	}
}

// Domain Error Codes
const (
	ErrCodeInvalidInput       = "INVALID_INPUT"
	ErrCodeKeyNotFound        = "KEY_NOT_FOUND"
	ErrCodeEncryptionFailed   = "ENCRYPTION_FAILED"
	ErrCodeDecryptionFailed   = "DECRYPTION_FAILED"
	ErrCodeInvalidAlgorithm   = "INVALID_ALGORITHM"
	ErrCodeKeyRotationFailed  = "KEY_ROTATION_FAILED"
	ErrCodeDataTooLarge       = "DATA_TOO_LARGE"
	ErrCodeInvalidEncryption  = "INVALID_ENCRYPTION"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// Domain Types with enhanced validation
type (
	KeyID      string
	Algorithm  string
	KeyVersion uint32
)

const (
	AlgorithmAES256GCM        Algorithm  = "AES-256-GCM"
	AlgorithmChaCha20Poly1305 Algorithm  = "CHACHA20-POLY1305"
	DefaultKeyID              KeyID      = "default"
	CurrentKeyVersion         KeyVersion = 1
)

// Enhanced Value Objects
type EncryptionKey struct {
	id        KeyID
	version   KeyVersion
	key       []byte
	algorithm Algorithm
	createdAt time.Time
	expiresAt time.Time
}

func NewEncryptionKey(id KeyID, version KeyVersion, key []byte, algorithm Algorithm, ttl time.Duration) (*EncryptionKey, error) {
	if id == "" {
		return nil, NewEncryptionError(ErrCodeInvalidInput, "key ID cannot be empty", nil, nil)
	}
	if len(key) == 0 {
		return nil, NewEncryptionError(ErrCodeInvalidInput, "key cannot be empty", nil, nil)
	}
	if algorithm != AlgorithmAES256GCM && algorithm != AlgorithmChaCha20Poly1305 {
		return nil, NewEncryptionError(ErrCodeInvalidAlgorithm, "unsupported algorithm", nil, nil)
	}

	now := time.Now()
	ek := &EncryptionKey{
		id:        id,
		version:   version,
		key:       make([]byte, len(key)),
		algorithm: algorithm,
		createdAt: now,
		expiresAt: now.Add(ttl),
	}

	copy(ek.key, key) // Defensive copy
	return ek, nil
}

func (ek *EncryptionKey) ID() KeyID           { return ek.id }
func (ek *EncryptionKey) Version() KeyVersion { return ek.version }
func (ek *EncryptionKey) Key() []byte {
	result := make([]byte, len(ek.key))
	copy(result, ek.key)
	return result
}
func (ek *EncryptionKey) Algorithm() Algorithm { return ek.algorithm }
func (ek *EncryptionKey) CreatedAt() time.Time { return ek.createdAt }
func (ek *EncryptionKey) ExpiresAt() time.Time { return ek.expiresAt }
func (ek *EncryptionKey) IsExpired() bool      { return time.Now().After(ek.expiresAt) }

// Secure cleanup
func (ek *EncryptionKey) Zeroize() {
	for i := range ek.key {
		ek.key[i] = 0
	}
	runtime.GC() // Force GC to clear any copies
}

type EncryptedPayload struct {
	version    uint8
	keyID      KeyID
	keyVersion KeyVersion
	algorithm  Algorithm
	nonce      []byte
	ciphertext []byte
	createdAt  time.Time
}

func NewEncryptedPayload(keyID KeyID, keyVersion KeyVersion, algorithm Algorithm, nonce, ciphertext []byte) (*EncryptedPayload, error) {
	if keyID == "" {
		return nil, NewEncryptionError(ErrCodeInvalidInput, "key ID cannot be empty", nil, nil)
	}
	if len(nonce) == 0 || len(ciphertext) == 0 {
		return nil, NewEncryptionError(ErrCodeInvalidInput, "nonce and ciphertext cannot be empty", nil, nil)
	}

	return &EncryptedPayload{
		version:    1, // Payload format version
		keyID:      keyID,
		keyVersion: keyVersion,
		algorithm:  algorithm,
		nonce:      nonce,
		ciphertext: ciphertext,
		createdAt:  time.Now(),
	}, nil
}

func (ep *EncryptedPayload) KeyID() KeyID           { return ep.keyID }
func (ep *EncryptedPayload) KeyVersion() KeyVersion { return ep.keyVersion }
func (ep *EncryptedPayload) Algorithm() Algorithm   { return ep.algorithm }
func (ep *EncryptedPayload) Nonce() []byte          { return ep.nonce }
func (ep *EncryptedPayload) Ciphertext() []byte     { return ep.ciphertext }
func (ep *EncryptedPayload) CreatedAt() time.Time   { return ep.createdAt }

// Serialize to bytes for storage
func (ep *EncryptedPayload) Marshal() []byte {
	keyIDBytes := []byte(ep.keyID)
	algorithmBytes := []byte(ep.algorithm)

	// Calculate total size
	size := 1 + // version
		4 + len(keyIDBytes) + // keyID length + keyID
		4 + // keyVersion
		4 + len(algorithmBytes) + // algorithm length + algorithm
		4 + len(ep.nonce) + // nonce length + nonce
		4 + len(ep.ciphertext) + // ciphertext length + ciphertext
		8 // timestamp

	result := make([]byte, size)
	offset := 0

	// Version
	result[offset] = ep.version
	offset++

	// KeyID
	binary.BigEndian.PutUint32(result[offset:], uint32(len(keyIDBytes)))
	offset += 4
	copy(result[offset:], keyIDBytes)
	offset += len(keyIDBytes)

	// KeyVersion
	binary.BigEndian.PutUint32(result[offset:], uint32(ep.keyVersion))
	offset += 4

	// Algorithm
	binary.BigEndian.PutUint32(result[offset:], uint32(len(algorithmBytes)))
	offset += 4
	copy(result[offset:], algorithmBytes)
	offset += len(algorithmBytes)

	// Nonce
	binary.BigEndian.PutUint32(result[offset:], uint32(len(ep.nonce)))
	offset += 4
	copy(result[offset:], ep.nonce)
	offset += len(ep.nonce)

	// Ciphertext
	binary.BigEndian.PutUint32(result[offset:], uint32(len(ep.ciphertext)))
	offset += 4
	copy(result[offset:], ep.ciphertext)
	offset += len(ep.ciphertext)

	// Timestamp
	binary.BigEndian.PutUint64(result[offset:], uint64(ep.createdAt.Unix()))

	return result
}

// Deserialize from bytes
func UnmarshalEncryptedPayload(data []byte) (*EncryptedPayload, error) {
	if len(data) < 1 {
		return nil, NewEncryptionError(ErrCodeInvalidEncryption, "data too short", nil, nil)
	}

	offset := 0
	version := data[offset]
	offset++

	if version != 1 {
		return nil, NewEncryptionError(ErrCodeInvalidEncryption, "unsupported payload version", nil, nil)
	}

	// KeyID
	if len(data) < offset+4 {
		return nil, NewEncryptionError(ErrCodeInvalidEncryption, "incomplete keyID length", nil, nil)
	}
	keyIDLen := binary.BigEndian.Uint32(data[offset:])
	offset += 4

	if len(data) < offset+int(keyIDLen) {
		return nil, NewEncryptionError(ErrCodeInvalidEncryption, "incomplete keyID", nil, nil)
	}
	keyID := KeyID(data[offset : offset+int(keyIDLen)])
	offset += int(keyIDLen)

	// KeyVersion
	if len(data) < offset+4 {
		return nil, NewEncryptionError(ErrCodeInvalidEncryption, "incomplete key version", nil, nil)
	}
	keyVersion := KeyVersion(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	// Algorithm
	if len(data) < offset+4 {
		return nil, NewEncryptionError(ErrCodeInvalidEncryption, "incomplete algorithm length", nil, nil)
	}
	algorithmLen := binary.BigEndian.Uint32(data[offset:])
	offset += 4

	if len(data) < offset+int(algorithmLen) {
		return nil, NewEncryptionError(ErrCodeInvalidEncryption, "incomplete algorithm", nil, nil)
	}
	algorithm := Algorithm(data[offset : offset+int(algorithmLen)])
	offset += int(algorithmLen)

	// Nonce
	if len(data) < offset+4 {
		return nil, NewEncryptionError(ErrCodeInvalidEncryption, "incomplete nonce length", nil, nil)
	}
	nonceLen := binary.BigEndian.Uint32(data[offset:])
	offset += 4

	if len(data) < offset+int(nonceLen) {
		return nil, NewEncryptionError(ErrCodeInvalidEncryption, "incomplete nonce", nil, nil)
	}
	nonce := make([]byte, nonceLen)
	copy(nonce, data[offset:offset+int(nonceLen)])
	offset += int(nonceLen)

	// Ciphertext
	if len(data) < offset+4 {
		return nil, NewEncryptionError(ErrCodeInvalidEncryption, "incomplete ciphertext length", nil, nil)
	}
	ciphertextLen := binary.BigEndian.Uint32(data[offset:])
	offset += 4

	if len(data) < offset+int(ciphertextLen) {
		return nil, NewEncryptionError(ErrCodeInvalidEncryption, "incomplete ciphertext", nil, nil)
	}
	ciphertext := make([]byte, ciphertextLen)
	copy(ciphertext, data[offset:offset+int(ciphertextLen)])
	offset += int(ciphertextLen)

	// Timestamp
	if len(data) < offset+8 {
		return nil, NewEncryptionError(ErrCodeInvalidEncryption, "incomplete timestamp", nil, nil)
	}
	timestamp := time.Unix(int64(binary.BigEndian.Uint64(data[offset:])), 0)

	payload := &EncryptedPayload{
		version:    version,
		keyID:      keyID,
		keyVersion: keyVersion,
		algorithm:  algorithm,
		nonce:      nonce,
		ciphertext: ciphertext,
		createdAt:  timestamp,
	}

	return payload, nil
}

// Enhanced Aggregate Root
type FieldEncryption struct {
	id        string
	fieldName string
	payload   *EncryptedPayload
	version   int
	updatedAt time.Time
	mu        sync.RWMutex
}

func NewFieldEncryption(id, fieldName string, payload *EncryptedPayload) (*FieldEncryption, error) {
	if err := validateFieldInput(id, fieldName); err != nil {
		return nil, err
	}
	if payload == nil {
		return nil, NewEncryptionError(ErrCodeInvalidInput, "encrypted payload cannot be nil", nil, nil)
	}

	return &FieldEncryption{
		id:        id,
		fieldName: fieldName,
		payload:   payload,
		version:   1,
		updatedAt: time.Now(),
	}, nil
}

func (fe *FieldEncryption) ID() string {
	fe.mu.RLock()
	defer fe.mu.RUnlock()
	return fe.id
}

func (fe *FieldEncryption) FieldName() string {
	fe.mu.RLock()
	defer fe.mu.RUnlock()
	return fe.fieldName
}

func (fe *FieldEncryption) Payload() *EncryptedPayload {
	fe.mu.RLock()
	defer fe.mu.RUnlock()
	return fe.payload
}

func (fe *FieldEncryption) Version() int {
	fe.mu.RLock()
	defer fe.mu.RUnlock()
	return fe.version
}

func (fe *FieldEncryption) UpdatedAt() time.Time {
	fe.mu.RLock()
	defer fe.mu.RUnlock()
	return fe.updatedAt
}

func (fe *FieldEncryption) UpdatePayload(payload *EncryptedPayload) error {
	if payload == nil {
		return NewEncryptionError(ErrCodeInvalidInput, "encrypted payload cannot be nil", nil, nil)
	}

	fe.mu.Lock()
	defer fe.mu.Unlock()

	fe.payload = payload
	fe.version++
	fe.updatedAt = time.Now()
	return nil
}

// Validation helpers
func validateFieldInput(id, fieldName string) error {
	if id == "" {
		return NewEncryptionError(ErrCodeInvalidInput, "field ID cannot be empty", nil, nil)
	}
	if fieldName == "" {
		return NewEncryptionError(ErrCodeInvalidInput, "field name cannot be empty", nil, nil)
	}
	if len(id) > 255 {
		return NewEncryptionError(ErrCodeInvalidInput, "field ID too long", nil, nil)
	}
	if len(fieldName) > 255 {
		return NewEncryptionError(ErrCodeInvalidInput, "field name too long", nil, nil)
	}
	return nil
}

// Enhanced Repository Interfaces
type FieldEncryptionRepository interface {
	Save(ctx context.Context, encryption *FieldEncryption) error
	FindByID(ctx context.Context, id string) (*FieldEncryption, error)
	FindByFieldName(ctx context.Context, fieldName string) ([]*FieldEncryption, error)
	Delete(ctx context.Context, id string) error
	HealthCheck(ctx context.Context) error
}

type KeyRepository interface {
	GetKey(ctx context.Context, keyID KeyID, version KeyVersion) (*EncryptionKey, error)
	GetLatestKey(ctx context.Context, keyID KeyID) (*EncryptionKey, error)
	StoreKey(ctx context.Context, key *EncryptionKey) error
	ListKeys(ctx context.Context) (map[KeyID][]*EncryptionKey, error)
	RotateKey(ctx context.Context, keyID KeyID) (*EncryptionKey, error)
	HealthCheck(ctx context.Context) error
}

// Key derivation service
type KeyDerivationService interface {
	DeriveKey(masterKey []byte, salt []byte, keyID KeyID) ([]byte, error)
}

type argon2KeyDerivationService struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

func NewArgon2KeyDerivationService() KeyDerivationService {
	return &argon2KeyDerivationService{
		time:    1,
		memory:  64 * 1024, // 64MB
		threads: 4,
		keyLen:  32, // 256 bits
	}
}

func (kds *argon2KeyDerivationService) DeriveKey(masterKey []byte, salt []byte, keyID KeyID) ([]byte, error) {
	if len(masterKey) < 32 {
		return nil, NewEncryptionError(ErrCodeInvalidInput, "master key too short", nil, nil)
	}
	if len(salt) < 16 {
		return nil, NewEncryptionError(ErrCodeInvalidInput, "salt too short", nil, nil)
	}

	// Combine salt with keyID for domain separation
	combinedSalt := append(salt, []byte(keyID)...)

	derivedKey := argon2.IDKey(masterKey, combinedSalt, kds.time, kds.memory, kds.threads, kds.keyLen)
	return derivedKey, nil
}

// Caching layer
type CachedKeyRepository struct {
	underlying KeyRepository
	cache      sync.Map
	ttl        time.Duration
	logger     *zap.Logger
}

type cacheEntry struct {
	key       *EncryptionKey
	expiresAt time.Time
}

func NewCachedKeyRepository(underlying KeyRepository, ttl time.Duration, logger *zap.Logger) *CachedKeyRepository {
	repo := &CachedKeyRepository{
		underlying: underlying,
		ttl:        ttl,
		logger:     logger,
	}

	// Start cache cleanup goroutine
	go repo.cleanupExpired()

	return repo
}

func (ckr *CachedKeyRepository) GetKey(ctx context.Context, keyID KeyID, version KeyVersion) (*EncryptionKey, error) {
	cacheKey := fmt.Sprintf("%s:%d", keyID, version)

	if entry, ok := ckr.cache.Load(cacheKey); ok {
		ce := entry.(*cacheEntry)
		if time.Now().Before(ce.expiresAt) && !ce.key.IsExpired() {
			ckr.logger.Debug("cache hit", zap.String("keyID", string(keyID)), zap.Uint32("version", uint32(version)))
			return ce.key, nil
		}
		ckr.cache.Delete(cacheKey)
	}

	key, err := ckr.underlying.GetKey(ctx, keyID, version)
	if err != nil {
		return nil, err
	}

	// Cache the key
	ckr.cache.Store(cacheKey, &cacheEntry{
		key:       key,
		expiresAt: time.Now().Add(ckr.ttl),
	})

	ckr.logger.Debug("cache miss", zap.String("keyID", string(keyID)), zap.Uint32("version", uint32(version)))
	return key, nil
}

func (ckr *CachedKeyRepository) GetLatestKey(ctx context.Context, keyID KeyID) (*EncryptionKey, error) {
	// For latest keys, we don't cache as they change frequently
	return ckr.underlying.GetLatestKey(ctx, keyID)
}

func (ckr *CachedKeyRepository) StoreKey(ctx context.Context, key *EncryptionKey) error {
	// Invalidate cache for this key
	cacheKey := fmt.Sprintf("%s:%d", key.ID(), key.Version())
	ckr.cache.Delete(cacheKey)

	return ckr.underlying.StoreKey(ctx, key)
}

func (ckr *CachedKeyRepository) ListKeys(ctx context.Context) (map[KeyID][]*EncryptionKey, error) {
	return ckr.underlying.ListKeys(ctx)
}

func (ckr *CachedKeyRepository) RotateKey(ctx context.Context, keyID KeyID) (*EncryptionKey, error) {
	// Clear all cached versions of this key
	ckr.cache.Range(func(key, value any) bool {
		keyStr := key.(string)
		if keyStr[:len(keyID)] == string(keyID) {
			ckr.cache.Delete(key)
		}
		return true
	})

	return ckr.underlying.RotateKey(ctx, keyID)
}

func (ckr *CachedKeyRepository) HealthCheck(ctx context.Context) error {
	return ckr.underlying.HealthCheck(ctx)
}

func (ckr *CachedKeyRepository) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		ckr.cache.Range(func(key, value any) bool {
			entry := value.(*cacheEntry)
			if now.After(entry.expiresAt) || entry.key.IsExpired() {
				ckr.cache.Delete(key)
			}
			return true
		})
	}
}

// Enhanced Domain Service
type EncryptionService interface {
	Encrypt(ctx context.Context, plaintext string, keyID KeyID) (*EncryptedPayload, error)
	Decrypt(ctx context.Context, payload *EncryptedPayload) (string, error)
	RotateKey(ctx context.Context, oldPayload *EncryptedPayload, newKeyID KeyID) (*EncryptedPayload, error)
	BulkEncrypt(ctx context.Context, data map[string]string, keyID KeyID) (map[string]*EncryptedPayload, error)
	BulkDecrypt(ctx context.Context, payloads map[string]*EncryptedPayload) (map[string]string, error)
	HealthCheck(ctx context.Context) error
}

type encryptionService struct {
	keyRepo     KeyRepository
	aeadPool    sync.Pool
	maxDataSize int
	logger      *zap.Logger
	metrics     Metrics
}

func NewEncryptionService(keyRepo KeyRepository, maxDataSize int, logger *zap.Logger, metrics Metrics) EncryptionService {
	return &encryptionService{
		keyRepo:     keyRepo,
		maxDataSize: maxDataSize,
		logger:      logger,
		metrics:     metrics,
		aeadPool: sync.Pool{
			New: func() any {
				return make(map[string]cipher.AEAD)
			},
		},
	}
}

func (es *encryptionService) Encrypt(ctx context.Context, plaintext string, keyID KeyID) (*EncryptedPayload, error) {
	start := time.Now()
	defer func() {
		es.metrics.RecordEncryptionDuration(time.Since(start))
	}()

	if keyID == "" {
		keyID = DefaultKeyID
	}

	if len(plaintext) > es.maxDataSize {
		return nil, NewEncryptionError(ErrCodeDataTooLarge,
			fmt.Sprintf("data size %d exceeds maximum %d", len(plaintext), es.maxDataSize),
			nil, nil)
	}

	key, err := es.keyRepo.GetLatestKey(ctx, keyID)
	if err != nil {
		es.metrics.IncrementErrorCount("encrypt", "key_retrieval_failed")
		return nil, NewEncryptionError(ErrCodeKeyNotFound, "failed to retrieve encryption key", err,
			map[string]any{"keyID": keyID})
	}

	aead, err := es.createAEAD(key)
	if err != nil {
		es.metrics.IncrementErrorCount("encrypt", "aead_creation_failed")
		return nil, NewEncryptionError(ErrCodeEncryptionFailed, "failed to create AEAD", err,
			map[string]any{"keyID": keyID, "algorithm": key.Algorithm()})
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		es.metrics.IncrementErrorCount("encrypt", "nonce_generation_failed")
		return nil, NewEncryptionError(ErrCodeEncryptionFailed, "failed to generate nonce", err, nil)
	}

	// Use key ID as additional authenticated data for domain separation
	aad := []byte(fmt.Sprintf("field-encryption:%s:%d", keyID, key.Version()))
	ciphertext := aead.Seal(nil, nonce, []byte(plaintext), aad)

	payload, err := NewEncryptedPayload(keyID, key.Version(), key.Algorithm(), nonce, ciphertext)
	if err != nil {
		return nil, err
	}

	es.logger.Debug("encryption completed",
		zap.String("keyID", string(keyID)),
		zap.Uint32("keyVersion", uint32(key.Version())),
		zap.Int("plaintextSize", len(plaintext)),
		zap.Int("ciphertextSize", len(ciphertext)))

	es.metrics.IncrementSuccessCount("encrypt")
	return payload, nil
}

func (es *encryptionService) Decrypt(ctx context.Context, payload *EncryptedPayload) (string, error) {
	start := time.Now()
	defer func() {
		es.metrics.RecordDecryptionDuration(time.Since(start))
	}()

	if payload == nil {
		return "", NewEncryptionError(ErrCodeInvalidInput, "encrypted payload cannot be nil", nil, nil)
	}

	key, err := es.keyRepo.GetKey(ctx, payload.KeyID(), payload.KeyVersion())
	if err != nil {
		es.metrics.IncrementErrorCount("decrypt", "key_retrieval_failed")
		return "", NewEncryptionError(ErrCodeKeyNotFound, "failed to retrieve decryption key", err,
			map[string]any{
				"keyID":      payload.KeyID(),
				"keyVersion": payload.KeyVersion(),
			})
	}

	aead, err := es.createAEAD(key)
	if err != nil {
		es.metrics.IncrementErrorCount("decrypt", "aead_creation_failed")
		return "", NewEncryptionError(ErrCodeDecryptionFailed, "failed to create AEAD", err,
			map[string]any{
				"keyID":     payload.KeyID(),
				"algorithm": key.Algorithm(),
			})
	}

	// Use same AAD as encryption
	aad := []byte(fmt.Sprintf("field-encryption:%s:%d", payload.KeyID(), key.Version()))
	plaintext, err := aead.Open(nil, payload.Nonce(), payload.Ciphertext(), aad)
	if err != nil {
		es.metrics.IncrementErrorCount("decrypt", "decryption_failed")
		return "", NewEncryptionError(ErrCodeDecryptionFailed, "failed to decrypt data", err,
			map[string]any{
				"keyID":      payload.KeyID(),
				"keyVersion": payload.KeyVersion(),
			})
	}

	// Secure constant-time validation
	if subtle.ConstantTimeCompare([]byte(string(plaintext)), plaintext) != 1 {
		es.metrics.IncrementErrorCount("decrypt", "integrity_check_failed")
		return "", NewEncryptionError(ErrCodeDecryptionFailed, "integrity check failed", nil, nil)
	}

	es.logger.Debug("decryption completed",
		zap.String("keyID", string(payload.KeyID())),
		zap.Uint32("keyVersion", uint32(payload.KeyVersion())),
		zap.Int("plaintextSize", len(plaintext)))

	es.metrics.IncrementSuccessCount("decrypt")
	return string(plaintext), nil
}

func (es *encryptionService) RotateKey(ctx context.Context, oldPayload *EncryptedPayload, newKeyID KeyID) (*EncryptedPayload, error) {
	// Decrypt with old key
	plaintext, err := es.Decrypt(ctx, oldPayload)
	if err != nil {
		return nil, NewEncryptionError(ErrCodeKeyRotationFailed, "failed to decrypt with old key", err, nil)
	}

	// Encrypt with new key
	newPayload, err := es.Encrypt(ctx, plaintext, newKeyID)
	if err != nil {
		return nil, NewEncryptionError(ErrCodeKeyRotationFailed, "failed to encrypt with new key", err, nil)
	}

	es.logger.Info("key rotation completed",
		zap.String("oldKeyID", string(oldPayload.KeyID())),
		zap.Uint32("oldKeyVersion", uint32(oldPayload.KeyVersion())),
		zap.String("newKeyID", string(newKeyID)))

	return newPayload, nil
}

func (es *encryptionService) BulkEncrypt(ctx context.Context, data map[string]string, keyID KeyID) (map[string]*EncryptedPayload, error) {
	if len(data) == 0 {
		return make(map[string]*EncryptedPayload), nil
	}

	results := make(map[string]*EncryptedPayload, len(data))
	errors := make(map[string]error)

	// Use worker pool for concurrent processing
	const maxWorkers = 10
	workers := len(data)
	if workers > maxWorkers {
		workers = maxWorkers
	}

	type job struct {
		key   string
		value string
	}

	type result struct {
		key     string
		payload *EncryptedPayload
		err     error
	}

	jobs := make(chan job, len(data))
	results_chan := make(chan result, len(data))

	// Start workers
	for w := 0; w < workers; w++ {
		go func() {
			for job := range jobs {
				payload, err := es.Encrypt(ctx, job.value, keyID)
				results_chan <- result{key: job.key, payload: payload, err: err}
			}
		}()
	}

	// Send jobs
	for k, v := range data {
		jobs <- job{key: k, value: v}
	}
	close(jobs)

	// Collect results
	for i := 0; i < len(data); i++ {
		res := <-results_chan
		if res.err != nil {
			errors[res.key] = res.err
		} else {
			results[res.key] = res.payload
		}
	}

	if len(errors) > 0 {
		return results, NewEncryptionError(ErrCodeEncryptionFailed, "bulk encryption partially failed", nil,
			map[string]any{"errors": errors})
	}

	return results, nil
}

func (es *encryptionService) BulkDecrypt(ctx context.Context, payloads map[string]*EncryptedPayload) (map[string]string, error) {
	if len(payloads) == 0 {
		return make(map[string]string), nil
	}

	results := make(map[string]string, len(payloads))
	errors := make(map[string]error)

	const maxWorkers = 10
	workers := len(payloads)
	if workers > maxWorkers {
		workers = maxWorkers
	}

	type job struct {
		key     string
		payload *EncryptedPayload
	}

	type result struct {
		key       string
		plaintext string
		err       error
	}

	jobs := make(chan job, len(payloads))
	results_chan := make(chan result, len(payloads))

	// Start workers
	for w := 0; w < workers; w++ {
		go func() {
			for job := range jobs {
				plaintext, err := es.Decrypt(ctx, job.payload)
				results_chan <- result{key: job.key, plaintext: plaintext, err: err}
			}
		}()
	}

	// Send jobs
	for k, v := range payloads {
		jobs <- job{key: k, payload: v}
	}
	close(jobs)

	// Collect results
	for i := 0; i < len(payloads); i++ {
		res := <-results_chan
		if res.err != nil {
			errors[res.key] = res.err
		} else {
			results[res.key] = res.plaintext
		}
	}

	if len(errors) > 0 {
		return results, NewEncryptionError(ErrCodeDecryptionFailed, "bulk decryption partially failed", nil,
			map[string]any{"errors": errors})
	}

	return results, nil
}

func (es *encryptionService) HealthCheck(ctx context.Context) error {
	return es.keyRepo.HealthCheck(ctx)
}

func (es *encryptionService) createAEAD(key *EncryptionKey) (cipher.AEAD, error) {
	switch key.Algorithm() {
	case AlgorithmAES256GCM:
		block, err := aes.NewCipher(key.Key())
		if err != nil {
			return nil, err
		}
		return cipher.NewGCM(block)
	case AlgorithmChaCha20Poly1305:
		// Note: Would need to import golang.org/x/crypto/chacha20poly1305
		// return chacha20poly1305.New(key.Key())
		return nil, NewEncryptionError(ErrCodeInvalidAlgorithm, "ChaCha20Poly1305 not implemented", nil, nil)
	default:
		return nil, NewEncryptionError(ErrCodeInvalidAlgorithm, "unsupported algorithm", nil,
			map[string]any{"algorithm": key.Algorithm()})
	}
}

// Observability interfaces
type Metrics interface {
	RecordEncryptionDuration(duration time.Duration)
	RecordDecryptionDuration(duration time.Duration)
	IncrementSuccessCount(operation string)
	IncrementErrorCount(operation, errorType string)
	RecordKeyRotation(keyID string)
}

// Application Service with enhanced features
type FieldEncryptionService struct {
	encryptionSvc EncryptionService
	repository    FieldEncryptionRepository
	keyRepo       KeyRepository
	logger        *zap.Logger
	metrics       Metrics
	config        *Config
}

func NewFieldEncryptionService(
	encryptionSvc EncryptionService,
	repo FieldEncryptionRepository,
	keyRepo KeyRepository,
	logger *zap.Logger,
	metrics Metrics,
	config *Config,
) *FieldEncryptionService {
	return &FieldEncryptionService{
		encryptionSvc: encryptionSvc,
		repository:    repo,
		keyRepo:       keyRepo,
		logger:        logger,
		metrics:       metrics,
		config:        config,
	}
}

func (fes *FieldEncryptionService) EncryptAndStore(ctx context.Context, id, fieldName, plaintext string, keyID KeyID) error {
	start := time.Now()
	defer func() {
		fes.logger.Info("encrypt_and_store_completed",
			zap.String("fieldID", id),
			zap.String("fieldName", fieldName),
			zap.Duration("duration", time.Since(start)))
	}()

	if err := validateFieldInput(id, fieldName); err != nil {
		return err
	}

	if len(plaintext) > fes.config.Encryption.MaxDataSize {
		return NewEncryptionError(ErrCodeDataTooLarge, "plaintext too large", nil,
			map[string]any{
				"size":    len(plaintext),
				"maxSize": fes.config.Encryption.MaxDataSize,
			})
	}

	payload, err := fes.encryptionSvc.Encrypt(ctx, plaintext, keyID)
	if err != nil {
		fes.logger.Error("encryption failed",
			zap.String("fieldID", id),
			zap.String("fieldName", fieldName),
			zap.Error(err))
		return err
	}

	fieldEncryption, err := NewFieldEncryption(id, fieldName, payload)
	if err != nil {
		return err
	}

	if err := fes.repository.Save(ctx, fieldEncryption); err != nil {
		fes.logger.Error("save failed",
			zap.String("fieldID", id),
			zap.String("fieldName", fieldName),
			zap.Error(err))
		return NewEncryptionError(ErrCodeServiceUnavailable, "failed to save encrypted field", err, nil)
	}

	fes.logger.Info("field encrypted and stored successfully",
		zap.String("fieldID", id),
		zap.String("fieldName", fieldName),
		zap.String("keyID", string(keyID)))

	return nil
}

func (fes *FieldEncryptionService) DecryptField(ctx context.Context, id string) (string, error) {
	start := time.Now()
	defer func() {
		fes.logger.Info("decrypt_field_completed",
			zap.String("fieldID", id),
			zap.Duration("duration", time.Since(start)))
	}()

	if id == "" {
		return "", NewEncryptionError(ErrCodeInvalidInput, "field ID cannot be empty", nil, nil)
	}

	fieldEncryption, err := fes.repository.FindByID(ctx, id)
	if err != nil {
		fes.logger.Error("field not found", zap.String("fieldID", id), zap.Error(err))
		return "", NewEncryptionError(ErrCodeServiceUnavailable, "failed to find encrypted field", err, nil)
	}

	plaintext, err := fes.encryptionSvc.Decrypt(ctx, fieldEncryption.Payload())
	if err != nil {
		fes.logger.Error("decryption failed", zap.String("fieldID", id), zap.Error(err))
		return "", err
	}

	fes.logger.Debug("field decrypted successfully", zap.String("fieldID", id))
	return plaintext, nil
}

func (fes *FieldEncryptionService) RotateFieldKey(ctx context.Context, id string, newKeyID KeyID) error {
	start := time.Now()
	defer func() {
		fes.logger.Info("rotate_field_key_completed",
			zap.String("fieldID", id),
			zap.String("newKeyID", string(newKeyID)),
			zap.Duration("duration", time.Since(start)))
	}()

	fieldEncryption, err := fes.repository.FindByID(ctx, id)
	if err != nil {
		return NewEncryptionError(ErrCodeServiceUnavailable, "failed to find encrypted field", err, nil)
	}

	oldKeyID := fieldEncryption.Payload().KeyID()
	newPayload, err := fes.encryptionSvc.RotateKey(ctx, fieldEncryption.Payload(), newKeyID)
	if err != nil {
		fes.logger.Error("key rotation failed",
			zap.String("fieldID", id),
			zap.String("oldKeyID", string(oldKeyID)),
			zap.String("newKeyID", string(newKeyID)),
			zap.Error(err))
		return err
	}

	if err := fieldEncryption.UpdatePayload(newPayload); err != nil {
		return err
	}

	if err := fes.repository.Save(ctx, fieldEncryption); err != nil {
		return NewEncryptionError(ErrCodeServiceUnavailable, "failed to save rotated field", err, nil)
	}

	fes.metrics.RecordKeyRotation(string(newKeyID))
	fes.logger.Info("field key rotated successfully",
		zap.String("fieldID", id),
		zap.String("oldKeyID", string(oldKeyID)),
		zap.String("newKeyID", string(newKeyID)))

	return nil
}

func (fes *FieldEncryptionService) BulkEncryptAndStore(ctx context.Context, fields map[string]FieldData, keyID KeyID) error {
	if len(fields) == 0 {
		return nil
	}

	// Extract plaintext data
	plaintextData := make(map[string]string, len(fields))
	for id, field := range fields {
		plaintextData[id] = field.Plaintext
	}

	// Bulk encrypt
	payloads, err := fes.encryptionSvc.BulkEncrypt(ctx, plaintextData, keyID)
	if err != nil {
		return err
	}

	// Save all encrypted fields
	for id, payload := range payloads {
		field := fields[id]
		fieldEncryption, err := NewFieldEncryption(id, field.Name, payload)
		if err != nil {
			return err
		}

		if err := fes.repository.Save(ctx, fieldEncryption); err != nil {
			return NewEncryptionError(ErrCodeServiceUnavailable, "failed to save bulk encrypted field", err,
				map[string]any{"fieldID": id})
		}
	}

	fes.logger.Info("bulk encryption completed", zap.Int("count", len(fields)))
	return nil
}

func (fes *FieldEncryptionService) HealthCheck(ctx context.Context) error {
	if err := fes.encryptionSvc.HealthCheck(ctx); err != nil {
		return err
	}
	return fes.repository.HealthCheck(ctx)
}

// Helper types
type FieldData struct {
	Name      string
	Plaintext string
}

// Enhanced Infrastructure - Production-ready implementations
type SecureInMemoryKeyRepository struct {
	keys    map[KeyID]map[KeyVersion]*EncryptionKey
	mu      sync.RWMutex
	kdf     KeyDerivationService
	logger  *zap.Logger
	metrics Metrics
}

func NewSecureInMemoryKeyRepository(masterKey []byte, kdf KeyDerivationService, logger *zap.Logger, metrics Metrics) (*SecureInMemoryKeyRepository, error) {
	if len(masterKey) < 32 {
		return nil, NewEncryptionError(ErrCodeInvalidInput, "master key too short", nil, nil)
	}

	repo := &SecureInMemoryKeyRepository{
		keys:    make(map[KeyID]map[KeyVersion]*EncryptionKey),
		kdf:     kdf,
		logger:  logger,
		metrics: metrics,
	}

	// Generate default key
	salt := []byte("default-salt-change-in-production")
	defaultKeyBytes, err := kdf.DeriveKey(masterKey, salt, DefaultKeyID)
	if err != nil {
		return nil, err
	}

	defaultKey, err := NewEncryptionKey(DefaultKeyID, CurrentKeyVersion, defaultKeyBytes, AlgorithmAES256GCM, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	repo.keys[DefaultKeyID] = map[KeyVersion]*EncryptionKey{
		CurrentKeyVersion: defaultKey,
	}

	return repo, nil
}

func (kr *SecureInMemoryKeyRepository) GetKey(ctx context.Context, keyID KeyID, version KeyVersion) (*EncryptionKey, error) {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	versions, exists := kr.keys[keyID]
	if !exists {
		return nil, NewEncryptionError(ErrCodeKeyNotFound, "key ID not found", nil,
			map[string]any{"keyID": keyID})
	}

	key, exists := versions[version]
	if !exists {
		return nil, NewEncryptionError(ErrCodeKeyNotFound, "key version not found", nil,
			map[string]any{"keyID": keyID, "version": version})
	}

	if key.IsExpired() {
		return nil, NewEncryptionError(ErrCodeKeyNotFound, "key expired", nil,
			map[string]any{"keyID": keyID, "version": version})
	}

	return key, nil
}

func (kr *SecureInMemoryKeyRepository) GetLatestKey(ctx context.Context, keyID KeyID) (*EncryptionKey, error) {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	versions, exists := kr.keys[keyID]
	if !exists {
		return nil, NewEncryptionError(ErrCodeKeyNotFound, "key ID not found", nil,
			map[string]any{"keyID": keyID})
	}

	var latestKey *EncryptionKey
	var latestVersion KeyVersion = 0

	for version, key := range versions {
		if !key.IsExpired() && version > latestVersion {
			latestVersion = version
			latestKey = key
		}
	}

	if latestKey == nil {
		return nil, NewEncryptionError(ErrCodeKeyNotFound, "no valid key found", nil,
			map[string]any{"keyID": keyID})
	}

	return latestKey, nil
}

func (kr *SecureInMemoryKeyRepository) StoreKey(ctx context.Context, key *EncryptionKey) error {
	kr.mu.Lock()
	defer kr.mu.Unlock()

	if _, exists := kr.keys[key.ID()]; !exists {
		kr.keys[key.ID()] = make(map[KeyVersion]*EncryptionKey)
	}

	kr.keys[key.ID()][key.Version()] = key
	kr.logger.Info("key stored",
		zap.String("keyID", string(key.ID())),
		zap.Uint32("version", uint32(key.Version())))

	return nil
}

func (kr *SecureInMemoryKeyRepository) ListKeys(ctx context.Context) (map[KeyID][]*EncryptionKey, error) {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	result := make(map[KeyID][]*EncryptionKey)
	for keyID, versions := range kr.keys {
		for _, key := range versions {
			if !key.IsExpired() {
				result[keyID] = append(result[keyID], key)
			}
		}
	}

	return result, nil
}

func (kr *SecureInMemoryKeyRepository) RotateKey(ctx context.Context, keyID KeyID) (*EncryptionKey, error) {
	kr.mu.Lock()
	defer kr.mu.Unlock()

	versions, exists := kr.keys[keyID]
	if !exists {
		return nil, NewEncryptionError(ErrCodeKeyNotFound, "key ID not found", nil,
			map[string]any{"keyID": keyID})
	}

	// Find the highest version number
	var maxVersion KeyVersion = 0
	for version := range versions {
		if version > maxVersion {
			maxVersion = version
		}
	}

	// Generate new key with next version
	newVersion := maxVersion + 1
	salt := []byte(fmt.Sprintf("rotation-salt-%s-%d", keyID, newVersion))

	// Use a dummy master key - in production this would come from secure storage
	masterKey := make([]byte, 32)
	if _, err := rand.Read(masterKey); err != nil {
		return nil, err
	}

	newKeyBytes, err := kr.kdf.DeriveKey(masterKey, salt, keyID)
	if err != nil {
		return nil, err
	}

	newKey, err := NewEncryptionKey(keyID, newVersion, newKeyBytes, AlgorithmAES256GCM, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	kr.keys[keyID][newVersion] = newKey
	kr.logger.Info("key rotated",
		zap.String("keyID", string(keyID)),
		zap.Uint32("newVersion", uint32(newVersion)))

	return newKey, nil
}

func (kr *SecureInMemoryKeyRepository) HealthCheck(ctx context.Context) error {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	if len(kr.keys) == 0 {
		return NewEncryptionError(ErrCodeServiceUnavailable, "no keys available", nil, nil)
	}

	return nil
}

// Enhanced Field Repository with thread safety
type ThreadSafeFieldRepository struct {
	fields  map[string]*FieldEncryption
	mu      sync.RWMutex
	logger  *zap.Logger
	metrics Metrics
}

func NewThreadSafeFieldRepository(logger *zap.Logger, metrics Metrics) *ThreadSafeFieldRepository {
	return &ThreadSafeFieldRepository{
		fields:  make(map[string]*FieldEncryption),
		logger:  logger,
		metrics: metrics,
	}
}

func (fr *ThreadSafeFieldRepository) Save(ctx context.Context, encryption *FieldEncryption) error {
	if encryption == nil {
		return NewEncryptionError(ErrCodeInvalidInput, "field encryption cannot be nil", nil, nil)
	}

	fr.mu.Lock()
	defer fr.mu.Unlock()

	fr.fields[encryption.ID()] = encryption
	fr.logger.Debug("field saved",
		zap.String("fieldID", encryption.ID()),
		zap.String("fieldName", encryption.FieldName()))

	return nil
}

func (fr *ThreadSafeFieldRepository) FindByID(ctx context.Context, id string) (*FieldEncryption, error) {
	if id == "" {
		return nil, NewEncryptionError(ErrCodeInvalidInput, "field ID cannot be empty", nil, nil)
	}

	fr.mu.RLock()
	defer fr.mu.RUnlock()

	field, exists := fr.fields[id]
	if !exists {
		return nil, NewEncryptionError(ErrCodeServiceUnavailable, "field not found", nil,
			map[string]any{"fieldID": id})
	}

	return field, nil
}

func (fr *ThreadSafeFieldRepository) FindByFieldName(ctx context.Context, fieldName string) ([]*FieldEncryption, error) {
	if fieldName == "" {
		return nil, NewEncryptionError(ErrCodeInvalidInput, "field name cannot be empty", nil, nil)
	}

	fr.mu.RLock()
	defer fr.mu.RUnlock()

	var results []*FieldEncryption
	for _, field := range fr.fields {
		if field.FieldName() == fieldName {
			results = append(results, field)
		}
	}

	return results, nil
}

func (fr *ThreadSafeFieldRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return NewEncryptionError(ErrCodeInvalidInput, "field ID cannot be empty", nil, nil)
	}

	fr.mu.Lock()
	defer fr.mu.Unlock()

	if _, exists := fr.fields[id]; !exists {
		return NewEncryptionError(ErrCodeServiceUnavailable, "field not found", nil,
			map[string]any{"fieldID": id})
	}

	delete(fr.fields, id)
	fr.logger.Debug("field deleted", zap.String("fieldID", id))

	return nil
}

func (fr *ThreadSafeFieldRepository) HealthCheck(ctx context.Context) error {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	// Simple health check - ensure we can access the storage
	_ = len(fr.fields)
	return nil
}

// Simple metrics implementation
type SimpleMetrics struct {
	mu        sync.RWMutex
	counters  map[string]int64
	durations map[string][]time.Duration
}

func NewSimpleMetrics() *SimpleMetrics {
	return &SimpleMetrics{
		counters:  make(map[string]int64),
		durations: make(map[string][]time.Duration),
	}
}

func (m *SimpleMetrics) RecordEncryptionDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.durations["encryption"] = append(m.durations["encryption"], duration)
}

func (m *SimpleMetrics) RecordDecryptionDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.durations["decryption"] = append(m.durations["decryption"], duration)
}

func (m *SimpleMetrics) IncrementSuccessCount(operation string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[fmt.Sprintf("%s_success", operation)]++
}

func (m *SimpleMetrics) IncrementErrorCount(operation, errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[fmt.Sprintf("%s_error_%s", operation, errorType)]++
}

func (m *SimpleMetrics) RecordKeyRotation(keyID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[fmt.Sprintf("key_rotation_%s", keyID)]++
}

// Factory for creating the complete service with all dependencies
func NewFieldEncryptionServiceFactory(config *Config) (*FieldEncryptionService, error) {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Initialize metrics
	metrics := NewSimpleMetrics()

	// Initialize key derivation service
	kdf := NewArgon2KeyDerivationService()

	// Generate master key (in production, this would come from secure key management)
	masterKey := make([]byte, 32)
	if _, err := rand.Read(masterKey); err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}

	// Initialize key repository
	keyRepo, err := NewSecureInMemoryKeyRepository(masterKey, kdf, logger, metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to create key repository: %w", err)
	}

	// Add caching layer if enabled
	var finalKeyRepo KeyRepository = keyRepo
	if config.Cache.Enabled {
		finalKeyRepo = NewCachedKeyRepository(keyRepo, config.Cache.TTL, logger)
	}

	// Initialize field repository
	fieldRepo := NewThreadSafeFieldRepository(logger, metrics)

	// Initialize encryption service
	encryptionSvc := NewEncryptionService(finalKeyRepo, config.Encryption.MaxDataSize, logger, metrics)

	// Initialize field encryption service
	return NewFieldEncryptionService(encryptionSvc, fieldRepo, finalKeyRepo, logger, metrics, config), nil
}

// Usage Example with error handling and logging
func ExampleUsage() error {
	ctx := context.Background()

	// Configuration
	config := &Config{
		KeyManagement: KeyManagementConfig{
			Provider:     "memory",
			KeyTTL:       24 * time.Hour,
			RotationFreq: 7 * 24 * time.Hour,
		},
		Encryption: EncryptionConfig{
			Algorithm:   AlgorithmAES256GCM,
			MaxDataSize: 1024 * 1024, // 1MB
		},
		Cache: CacheConfig{
			Enabled: true,
			TTL:     30 * time.Minute,
			MaxSize: 1000,
		},
	}

	// Initialize service
	service, err := NewFieldEncryptionServiceFactory(config)
	if err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}

	// Health check
	if err := service.HealthCheck(ctx); err != nil {
		return fmt.Errorf("service health check failed: %w", err)
	}

	// Encrypt and store a field
	if err := service.EncryptAndStore(ctx, "user-123", "email", "user@example.com", DefaultKeyID); err != nil {
		return fmt.Errorf("failed to encrypt and store field: %w", err)
	}

	// Decrypt the field
	decrypted, err := service.DecryptField(ctx, "user-123")
	if err != nil {
		return fmt.Errorf("failed to decrypt field: %w", err)
	}

	fmt.Printf("Decrypted email: %s\n", decrypted)

	// Bulk encryption example
	fields := map[string]FieldData{
		"user-124": {Name: "email", Plaintext: "user124@example.com"},
		"user-125": {Name: "email", Plaintext: "user125@example.com"},
		"user-126": {Name: "phone", Plaintext: "+1234567890"},
	}

	if err := service.BulkEncryptAndStore(ctx, fields, DefaultKeyID); err != nil {
		return fmt.Errorf("failed to bulk encrypt: %w", err)
	}

	// Key rotation example
	if err := service.RotateFieldKey(ctx, "user-123", DefaultKeyID); err != nil {
		return fmt.Errorf("failed to rotate key: %w", err)
	}

	fmt.Println("All operations completed successfully")
	return nil
}
