package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/gwaylib/errors"
)

const (
	_root_file = "storage_root.pem"
	_crt_file  = "storage_crt.pem"
	_key_file  = "storage_key.pem"
)

type CustomTLSCert struct {
	CertPath string
}

func NewCustomTLSCert(certPath string) (*CustomTLSCert, error) {
	// check root cert
	return &CustomTLSCert{CertPath: certPath}, nil
}

func (crt *CustomTLSCert) RootFile() string {
	return filepath.Join(crt.CertPath, _root_file)
}

func (crt *CustomTLSCert) CertFile() string {
	return filepath.Join(crt.CertPath, _crt_file)
}
func (crt *CustomTLSCert) KeyFile() string {
	return filepath.Join(crt.CertPath, _key_file)
}

func CreateTLSCert(certPath, keyPath string, ipAddrs []net.IP) error {
	max := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, max)
	subject := pkix.Name{
		Organization:       []string{"lib10"},
		OrganizationalUnit: []string{"bcstorage"},
		CommonName:         "auto-tls",
	}

	rootTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(100, 0, 0),
		KeyUsage:     x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  ipAddrs,
	}
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.As(err)
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, &pk.PublicKey, pk)
	if err != nil {
		return errors.As(err)
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		return errors.As(err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return errors.As(err)
	}

	certOut.Close()

	keyOut, err := os.Create(keyPath)
	if err != nil {
		return errors.As(err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)}); err != nil {
		return errors.As(err)
	}
	keyOut.Close()
	return nil
}
