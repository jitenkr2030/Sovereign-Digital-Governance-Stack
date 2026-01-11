package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"neam-platform/integration-gateway/config"
)

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Certificate *tls.Certificate
	CertPool    *x509.CertPool
	MinVersion  uint16
}

// LoadTLSConfig loads TLS configuration from config
func LoadTLSConfig(cfg *config.Config) (*tls.Config, error) {
	if !cfg.TLS.Enabled {
		// Return insecure config if TLS is disabled
		return &tls.Config{
			MinVersion: tls.VersionTLS12,
		}, nil
	}

	// Load server certificate
	cert, err := tls.LoadX509KeyPair(cfg.TLS.CertPath, cfg.TLS.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	// Load CA certificate for client verification
	var certPool *x509.CertPool
	if cfg.TLS.ClientAuthEnabled {
		caData, err := ioutil.ReadFile(cfg.TLS.CAPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		certPool = x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
	}

	// Parse min version
	minVersion := parseTLSVersion(cfg.TLS.MinVersion)
	if minVersion == 0 {
		minVersion = tls.VersionTLS12
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.NoClientCert,
		MinVersion:   minVersion,
	}

	// Enable mTLS if client auth is required
	if cfg.TLS.ClientAuthEnabled {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = certPool
	}

	return tlsConfig, nil
}

// CreateServerTLSConfig creates TLS configuration for server
func CreateServerTLSConfig(certPath, keyPath, caPath string, requireClientCert bool) (*tls.Config, error) {
	// Load server certificate
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	var certPool *x509.CertPool
	if caPath != "" {
		caData, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		certPool = x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	if requireClientCert && certPool != nil {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = certPool
	}

	return tlsConfig, nil
}

// CreateClientTLSConfig creates TLS configuration for client
func CreateClientTLSConfig(certPath, keyPath, caPath string) (*tls.Config, error) {
	var certs []tls.Certificate
	var certPool *x509.CertPool

	// Load client certificate if provided
	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		certs = append(certs, cert)
	}

	// Load CA certificate for server verification
	if caPath != "" {
		caData, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		certPool = x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
	}

	return &tls.Config{
		Certificates: certs,
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// CertificateManager manages certificates
type CertificateManager struct {
	caCert     *x509.Certificate
	caKey      *rsa.PrivateKey
	certDir    string
	validity   time.Duration
	keySize    int
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(certDir string, validity time.Duration) (*CertificateManager, error) {
	// Ensure directory exists
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create certificate directory: %w", err)
	}

	return &CertificateManager{
		certDir:  certDir,
		validity: validity,
		keySize:  2048,
	}, nil
}

// InitializeCA initializes a new CA certificate
func (cm *CertificateManager) InitializeCA() error {
	// Generate CA key
	caKey, err := rsa.GenerateKey(rand.Reader, cm.keySize)
	if err != nil {
		return fmt.Errorf("failed to generate CA key: %w", err)
	}

	// Create CA certificate
	caCert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"NEAM CA"},
			Country:       []string{"IN"},
			Province:      []string{""},
			Locality:      []string{"New Delhi"},
			StreetAddress: []string{""},
			CommonName:    "NEAM Integration Gateway CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(cm.validity),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}

	// Create self-signed CA certificate
	caCertBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, &caKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Save CA certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertBytes,
	})
	if err := ioutil.WriteFile(filepath.Join(cm.certDir, "ca.crt"), certPEM, 0644); err != nil {
		return fmt.Errorf("failed to save CA certificate: %w", err)
	}

	// Save CA key
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caKey),
	})
	if err := ioutil.WriteFile(filepath.Join(cm.certDir, "ca.key"), keyPEM, 0600); err != nil {
		return fmt.Errorf("failed to save CA key: %w", err)
	}

	cm.caCert = caCert
	cm.caKey = caKey

	return nil
}

// LoadCA loads existing CA certificate
func (cm *CertificateManager) LoadCA() error {
	// Load CA certificate
	certPEM, err := ioutil.ReadFile(filepath.Join(cm.certDir, "ca.crt"))
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return fmt.Errorf("invalid CA certificate format")
	}

	cm.caCert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Load CA key
	keyPEM, err := ioutil.ReadFile(filepath.Join(cm.certDir, "ca.key"))
	if err != nil {
		return fmt.Errorf("failed to read CA key: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil || keyBlock.Type != "RSA PRIVATE KEY" {
		return fmt.Errorf("invalid CA key format")
	}

	cm.caKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA key: %w", err)
	}

	return nil
}

// IssueCertificate issues a new certificate
func (cm *CertificateManager) IssueCertificate(cn string, org string, isCA bool) ([]byte, []byte, error) {
	if cm.caCert == nil || cm.caKey == nil {
		return nil, nil, fmt.Errorf("CA not initialized")
	}

	// Generate key
	key, err := rsa.GenerateKey(rand.Reader, cm.keySize)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Create certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization:  []string{org},
			Country:       []string{"IN"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			CommonName:    cn,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(cm.validity),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}

	if isCA {
		cert.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
		cert.MaxPathLen = 1
	}

	// Sign certificate with CA
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cm.caCert, &key.PublicKey, cm.caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	// Encode key
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	return certPEM, keyPEM, nil
}

// GetCACertificate returns the CA certificate
func (cm *CertificateManager) GetCACertificate() []byte {
	if cm.caCert == nil {
		return nil
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cm.caCert.Raw,
	})
}

// ValidateCertificate validates a certificate
func ValidateCertificate(certPEM []byte, caCertPEM []byte) (bool, error) {
	// Parse CA certificate
	caBlock, _ := pem.Decode(caCertPEM)
	if caBlock == nil {
		return false, fmt.Errorf("invalid CA certificate")
	}

	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Parse client certificate
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return false, fmt.Errorf("invalid certificate")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Verify certificate
	_, err = cert.Verify(x509.VerifyOptions{
		Roots:         nil,
		Intermediates: nil,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	})
	if err != nil {
		return false, fmt.Errorf("certificate verification failed: %w", err)
	}

	return true, nil
}

// parseTLSVersion parses TLS version string
func parseTLSVersion(version string) uint16 {
	switch version {
	case "1.0", "tls1.0", "TLSv1.0":
		return tls.VersionTLS10
	case "1.1", "tls1.1", "TLSv1.1":
		return tls.VersionTLS11
	case "1.2", "tls1.2", "TLSv1.2":
		return tls.VersionTLS12
	case "1.3", "tls1.3", "TLSv1.3":
		return tls.VersionTLS13
	default:
		return 0
	}
}
