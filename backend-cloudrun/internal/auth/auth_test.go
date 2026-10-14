package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testProject = "test-proj"

type fixture struct {
	key *rsa.PrivateKey
	jwk *httptest.Server
	v   *Verifier
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kid": "k1",
			"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		}}})
	}))
	t.Cleanup(srv.Close)
	return &fixture{key: key, jwk: srv, v: NewVerifier(testProject, srv.URL)}
}

func (f *fixture) token(t *testing.T, mutate func(map[string]any)) string {
	t.Helper()
	now := time.Now()
	payload := map[string]any{
		"iss": "https://securetoken.google.com/" + testProject, "aud": testProject, "sub": "u1",
		"iat": now.Add(-time.Minute).Unix(), "exp": now.Add(time.Hour).Unix(), "email": "a@b.co",
	}
	if mutate != nil {
		mutate(payload)
	}
	hb, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "k1"})
	pb, _ := json.Marshal(payload)
	signing := base64.RawURLEncoding.EncodeToString(hb) + "." + base64.RawURLEncoding.EncodeToString(pb)
	d := sha256.Sum256([]byte(signing))
	sig, err := rsa.SignPKCS1v15(rand.Reader, f.key, crypto.SHA256, d[:])
	if err != nil {
		t.Fatal(err)
	}
	return signing + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func TestVerify(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	c, err := f.v.Verify(ctx, f.token(t, nil))
	if err != nil || c.UID != "u1" || c.Role != "CITIZEN" {
		t.Fatalf("valid token: %+v %v", c, err)
	}
	c, err = f.v.Verify(ctx, f.token(t, func(p map[string]any) { p["role"] = "OFFICIAL" }))
	if err != nil || c.Role != "OFFICIAL" {
		t.Fatalf("role claim: %+v %v", c, err)
	}
	c, _ = f.v.Verify(ctx, f.token(t, func(p map[string]any) { p["role"] = "ROOT" }))
	if c == nil || c.Role != "CITIZEN" {
		t.Fatalf("unknown role must default to CITIZEN: %+v", c)
	}

	bad := map[string]func(map[string]any){
		"expired":  func(p map[string]any) { p["exp"] = time.Now().Add(-time.Hour).Unix() },
		"audience": func(p map[string]any) { p["aud"] = "other" },
		"issuer":   func(p map[string]any) { p["iss"] = "https://evil" },
		"subject":  func(p map[string]any) { p["sub"] = "" },
	}
	for name, m := range bad {
		if _, err := f.v.Verify(ctx, f.token(t, m)); err == nil {
			t.Errorf("%s should be rejected", name)
		}
	}
	tok := f.token(t, nil)
	if _, err := f.v.Verify(ctx, tok[:len(tok)-4]+"AAAA"); err == nil {
		t.Error("tampered signature should be rejected")
	}
	if _, err := f.v.Verify(ctx, "garbage"); err == nil {
		t.Error("malformed should be rejected")
	}
}

func TestMiddlewareAndRoles(t *testing.T) {
	f := newFixture(t)
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := Middleware(f.v, RequireRole(f.v, ok, "OFFICIAL", "ADMIN"))

	do := func(header string) int {
		req := httptest.NewRequest("POST", "/x", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w.Code
	}
	if got := do(""); got != 401 {
		t.Errorf("anonymous on guarded route: %d", got)
	}
	if got := do("Bearer nonsense"); got != 401 {
		t.Errorf("invalid token: %d", got)
	}
	if got := do("Bearer " + f.token(t, nil)); got != 403 {
		t.Errorf("citizen: %d", got)
	}
	if got := do("Bearer " + f.token(t, func(p map[string]any) { p["role"] = "ADMIN" })); got != 200 {
		t.Errorf("admin: %d", got)
	}
}

func TestDisabledVerifierIsOpen(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := Middleware(nil, RequireRole(nil, ok, "ADMIN"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/x", nil))
	if w.Code != 200 {
		t.Errorf("disabled verifier must not block: %d", w.Code)
	}
}
