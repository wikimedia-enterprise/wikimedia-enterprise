package creds

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"wikimedia-enterprise/services/bulk-ingestion/config/env"

	"google.golang.org/grpc/credentials"
)

// ErrFailedToAddCA identifies error when failing to add client certificate.
var ErrFailedToAddCA = errors.New("failed to add client CA's certificate")

// New creates new gRPC transport credentials that is used for MTLS.
func New(env *env.Environment) (credentials.TransportCredentials, error) {
	pol := x509.NewCertPool()

	if !pol.AppendCertsFromPEM([]byte(env.InternalRootCAPem)) {
		return nil, ErrFailedToAddCA
	}

	crt, err := tls.X509KeyPair([]byte(env.TLSCertPem), []byte(env.TLSPrivateKeyPem))

	if err != nil {
		return nil, err
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{crt},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pol,
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(cfg), nil
}
