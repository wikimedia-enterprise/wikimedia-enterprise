package creds_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"testing"
	"wikimedia-enterprise/services/bulk-ingestion/config/env"
	"wikimedia-enterprise/services/bulk-ingestion/packages/creds"

	"github.com/stretchr/testify/suite"
)

type certTestSuite struct {
	suite.Suite
	env *env.Environment
	err error
	cap string
	crp string
	pkp string
	rer bool // indicates that return value should be an error but we don't know what error
}

func (s *certTestSuite) SetupSuite() {
	s.env = &env.Environment{
		InternalRootCAPem: s.cap,
		TLSCertPem:        s.crp,
		TLSPrivateKeyPem:  s.pkp,
	}
}

func (s *certTestSuite) TestNew() {
	crd, err := creds.New(s.env)

	if s.err != nil {
		s.Assert().Equal(s.err, err)
	}

	if s.err == nil && !s.rer {
		s.Assert().NotNil(crd)
	}

	if s.rer {
		s.Assert().Error(err)
	}
}

func getCert(cpk *rsa.PrivateKey, cfg *x509.Certificate) (string, error) {
	crt, err := x509.CreateCertificate(rand.Reader, cfg, cfg, &cpk.PublicKey, cpk)

	if err != nil {
		return "", err
	}

	cpm := new(bytes.Buffer)

	if err := pem.Encode(cpm, &pem.Block{Type: "CERTIFICATE", Bytes: crt}); err != nil {
		return "", err
	}

	return cpm.String(), nil
}

func getCerts() (string, string, string, error) {
	cpk, err := rsa.GenerateKey(rand.Reader, 4096)

	if err != nil {
		return "", "", "", err
	}

	cfg := &x509.Certificate{
		SerialNumber: big.NewInt(1024),
	}

	cat, err := getCert(cpk, cfg)

	if err != nil {
		return "", "", "", err
	}

	crt, err := getCert(cpk, cfg)

	if err != nil {
		return "", "", "", err
	}

	pkk := new(bytes.Buffer)

	if err := pem.Encode(pkk, &pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(cpk)}); err != nil {
		return "", "", "", err
	}

	return cat, crt, pkk.String(), nil
}

func TestContainer(t *testing.T) {
	cap, crp, pkp, err := getCerts()

	if err != nil {
		t.Fatal(err)
	}

	for _, testCase := range []*certTestSuite{
		{
			cap: cap,
			crp: crp,
			pkp: pkp,
		},
		{
			crp: crp,
			pkp: pkp,
			err: creds.ErrFailedToAddCA,
		},
		{
			cap: cap,
			rer: true,
		},
		{
			cap: cap,
			crp: crp,
			rer: true,
		},
	} {
		suite.Run(t, testCase)
	}
}
