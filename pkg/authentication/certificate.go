package authentication

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Certificate - create new Certificate.
type Certificate interface {
	PublicKey() *rsa.PublicKey
	PrivateKey() *rsa.PrivateKey
	TLSCert() tls.Certificate
	DN() (string, error)
}

// certificate implements Certificate
type certificate struct {
	publicKey     *rsa.PublicKey
	privateKey    *rsa.PrivateKey
	tlsCert       tls.Certificate
	publicCertPem []byte
}

// NewCertificate - create new Certificate.
//
// Parameters:
// * publicKeyPem=PEM encoded public key.
// * privateKeyPem=PEM encoded private key.
//
// Returns Certificate, or nil with error set if something is invalid.
func NewCertificate(publicKeyPem, privateKeyPem string) (Certificate, error) {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPem))
	if err != nil {
		return nil, errors.Wrap(err, "error with public key")
	}
	publicPem := []byte(publicKeyPem)

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPem))
	if err != nil {
		return nil, errors.Wrap(err, "error with private key")
	}

	tlsCert, err := tls.X509KeyPair([]byte(publicKeyPem), []byte(privateKeyPem))
	if err != nil {
		logrus.StandardLogger().Warnln("tls.X509KeyPair, err=", err)
	}

	if err := validateKeys(publicKey, privateKey); err != nil {
		return nil, err
	}

	return &certificate{
		publicKey:     publicKey,
		privateKey:    privateKey,
		tlsCert:       tlsCert,
		publicCertPem: publicPem,
	}, nil
}

func (c certificate) PublicKey() *rsa.PublicKey {
	return c.publicKey
}

func (c certificate) PrivateKey() *rsa.PrivateKey {
	return c.privateKey
}

func (c certificate) TLSCert() tls.Certificate {
	return c.tlsCert
}

func validateKeys(publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey) error {
	// validate public and private key pair
	// see:
	// * https://stackoverflow.com/questions/20655702/signing-and-decoding-with-rsa-sha-in-go
	// * http://play.golang.org/p/bzpD7Pa9mr
	plaintext := []byte(`date: Thu, 05 Jan 2012 21:31:40 GMT`)

	hashed := sha256.Sum256(plaintext)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return errors.Wrap(err, "error signing")
	}

	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature); err != nil {
		return errors.Wrap(err, "error verifying")
	}

	return nil
}

func (c certificate) DN() (string, error) {

	cpb, _ := pem.Decode(c.publicCertPem)
	crt, err := x509.ParseCertificate(cpb.Bytes)
	if err != nil {
		logrus.Errorf("cannot parse cert %s", err.Error())
		return "", err
	}
	subject := crt.Subject

	var co string
	if len(subject.Country) > 0 {
		co = subject.Country[0]
	}

	var o string
	if len(subject.Organization) > 0 {
		o = subject.Organization[0]
	}

	var ou string
	if len(subject.OrganizationalUnit) > 0 {
		ou = subject.OrganizationalUnit[0]
	}

	cn := subject.CommonName
	dn := fmt.Sprintf("C=%s, O=%s, OU=%s, CN=%s", co, o, ou, cn)

	return dn, nil
}
