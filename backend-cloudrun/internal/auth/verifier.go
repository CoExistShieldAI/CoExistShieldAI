// Package auth verifies Firebase Auth ID tokens (RS256 JWTs signed by Google securetoken)
// using only the standard library.
package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

const DefaultJWKSURL = "https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com"

type Claims struct {
	UID   string
	Email string
	Role  string // CITIZEN | OFFICIAL | ADMIN (defaults to CITIZEN)
}

type Verifier struct {
	projectID string
	jwksURL   string
	client    *http.Client
	now       func() time.Time

	mu      sync.Mutex
	keys    map[string]*rsa.PublicKey
	expires time.Time
}

func NewVerifier(projectID, jwksURL string) *Verifier {
	if jwksURL == "" {
		jwksURL = DefaultJWKSURL
	}
	return &Verifier{
		projectID: projectID,
		jwksURL:   jwksURL,
		client:    &http.Client{Timeout: 5 * time.Second},
		now:       time.Now,
	}
}

var ErrInvalidToken = errors.New("invalid token")

func (v *Verifier) Verify(ctx context.Context, token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("%w: malformed", ErrInvalidToken)
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := decodeSegment(parts[0], &header); err != nil || header.Alg != "RS256" || header.Kid == "" {
		return nil, fmt.Errorf("%w: bad header", ErrInvalidToken)
	}
	var p struct {
		Iss   string `json:"iss"`
		Aud   string `json:"aud"`
		Sub   string `json:"sub"`
		Exp   int64  `json:"exp"`
		Iat   int64  `json:"iat"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := decodeSegment(parts[1], &p); err != nil {
		return nil, fmt.Errorf("%w: bad payload", ErrInvalidToken)
	}

	key, err := v.key(ctx, header.Kid)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("%w: bad signature encoding", ErrInvalidToken)
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], sig); err != nil {
		return nil, fmt.Errorf("%w: signature mismatch", ErrInvalidToken)
	}

	now := v.now()
	switch {
	case p.Aud != v.projectID:
		return nil, fmt.Errorf("%w: audience", ErrInvalidToken)
	case p.Iss != "https://securetoken.google.com/"+v.projectID:
		return nil, fmt.Errorf("%w: issuer", ErrInvalidToken)
	case p.Sub == "":
		return nil, fmt.Errorf("%w: subject", ErrInvalidToken)
	case p.Exp <= now.Unix():
		return nil, fmt.Errorf("%w: expired", ErrInvalidToken)
	case p.Iat > now.Add(5*time.Minute).Unix():
		return nil, fmt.Errorf("%w: issued in future", ErrInvalidToken)
	}

	role := "CITIZEN"
	if p.Role == "OFFICIAL" || p.Role == "ADMIN" {
		role = p.Role
	}
	return &Claims{UID: p.Sub, Email: p.Email, Role: role}, nil
}

func decodeSegment(seg string, out any) error {
	b, err := base64.RawURLEncoding.DecodeString(seg)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func (v *Verifier) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if k, ok := v.keys[kid]; ok && v.now().Before(v.expires) {
		return k, nil
	}
	if err := v.refreshLocked(ctx); err != nil {
		return nil, err
	}
	k, ok := v.keys[kid]
	if !ok {
		return nil, errors.New("unknown key id")
	}
	return k, nil
}

func (v *Verifier) refreshLocked(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks fetch status %d", resp.StatusCode)
	}
	var set struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return err
	}
	keys := map[string]*rsa.PublicKey{}
	for _, k := range set.Keys {
		nb, err1 := base64.RawURLEncoding.DecodeString(k.N)
		eb, err2 := base64.RawURLEncoding.DecodeString(k.E)
		if err1 != nil || err2 != nil {
			continue
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: int(new(big.Int).SetBytes(eb).Int64())}
	}
	if len(keys) == 0 {
		return errors.New("jwks contained no usable keys")
	}
	v.keys = keys
	v.expires = v.now().Add(time.Hour)
	return nil
}
