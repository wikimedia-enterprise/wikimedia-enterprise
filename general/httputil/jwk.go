package httputil

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
)

// JWKFetcher is an interface that wraps Fetch method.
type JWKFetcher interface {
	Fetch(interface{}) error
}

// JWKFinder is an interface that wraps Find method.
type JWKFinder interface {
	Find(string) (*JWK, error)
}

// JWKProvider is an interface that wraps JWKFetcher and JWKFinder interfaces.
type JWKProvider interface {
	JWKFetcher
	JWKFinder
}

// JWKWellKnown JSON web keys list.
type JWKWellKnown struct {
	Keys []*JWK `json:"keys"`
	mut  sync.Mutex
}

// Fetch get keys from the source.
func (j *JWKWellKnown) Fetch(iss interface{}) error {
	if len(j.Keys) > 0 {
		return nil
	}

	j.mut.Lock()
	defer j.mut.Unlock()

	res, err := http.Get(fmt.Sprintf("%s%s", iss, "/.well-known/jwks.json"))

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}

	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(j)
}

// Find key by identifier.
func (j *JWKWellKnown) Find(kid string) (*JWK, error) {
	for _, key := range j.Keys {
		if key.KID == kid {
			return key, nil
		}
	}

	return nil, errors.New("key not found")
}

// JWK public key meta data.
type JWK struct {
	Alg    string `json:"alg"`
	E      string `json:"e"`
	KID    string `json:"kid"`
	KTY    string `json:"kty"`
	N      string `json:"n"`
	Use    string `json:"use"`
	mut    sync.Mutex
	rsa256 *rsa.PublicKey
}

// RSA256 convert payload to a valid public key.
func (k *JWK) RSA256() (*rsa.PublicKey, error) {
	if k.rsa256 != nil {
		return k.rsa256, nil
	}

	k.mut.Lock()
	defer k.mut.Unlock()

	edc, err := base64.RawURLEncoding.DecodeString(k.E)

	if err != nil {
		return nil, err
	}

	if len(edc) < 4 {
		ndata := make([]byte, 4)
		copy(ndata[4-len(edc):], edc)
		edc = ndata
	}

	pub := &rsa.PublicKey{
		N: &big.Int{},
		E: int(binary.BigEndian.Uint32(edc[:])),
	}

	dcn, err := base64.RawURLEncoding.DecodeString(k.N)

	if err != nil {
		return nil, err
	}

	pub.N.SetBytes(dcn)
	k.rsa256 = pub

	return pub, nil
}
