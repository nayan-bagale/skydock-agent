package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/nayan-bagale/skydock-agent/internal/api"
	"golang.org/x/sync/singleflight"
)

const pendingLoginTTL = 2 * time.Minute
const refreshSkew = 60 * time.Second
const keepaliveInterval = 30 * time.Second

// OnChange is invoked when signed-in state or access token changes (login, logout, refresh).
type OnChange func(authenticated bool, accessToken string)

type pendingLogin struct {
	verifier  string
	state     string
	startedAt time.Time
}

// Session holds PKCE login state and API tokens for the desktop agent.
type Session struct {
	mu sync.Mutex

	api       *api.Client
	keys      KeyStore
	webOrigin string
	onChange  OnChange

	pending *pendingLogin

	signedIn     bool
	accessToken  string
	accessExpiry time.Time
	refreshToken string

	refreshGroup singleflight.Group
}

func New(apiClient *api.Client, webOrigin string, keys KeyStore, onChange OnChange) *Session {
	if onChange == nil {
		onChange = func(bool, string) {}
	}
	return &Session{
		api:       apiClient,
		keys:      keys,
		webOrigin: stringsTrimRightSlash(webOrigin),
		onChange:  onChange,
	}
}

func stringsTrimRightSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

// Start begins a browser login and returns the URL to open.
func (s *Session) Start(ctx context.Context) (string, error) {
	verifier, challenge, state, err := newPKCEPair()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.pending = &pendingLogin{
		verifier:  verifier,
		state:     state,
		startedAt: time.Now(),
	}
	s.mu.Unlock()

	u, err := url.Parse(s.webOrigin + "/login")
	if err != nil {
		return "", fmt.Errorf("login url: %w", err)
	}
	q := u.Query()
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// Complete finishes login after the desktop receives skydock://callback.
func (s *Session) Complete(ctx context.Context, code, state string) error {
	if code == "" {
		return errors.New("sign-in failed")
	}

	s.mu.Lock()
	pending := s.pending
	if pending == nil || time.Since(pending.startedAt) > pendingLoginTTL {
		s.pending = nil
		s.mu.Unlock()
		return errors.New("sign-in expired")
	}
	if len(state) != len(pending.state) ||
		subtle.ConstantTimeCompare([]byte(state), []byte(pending.state)) != 1 {
		s.pending = nil
		s.mu.Unlock()
		return errors.New("sign-in failed")
	}
	verifier := pending.verifier
	s.pending = nil
	s.mu.Unlock()

	tokens, err := s.api.PKCEExchange(ctx, api.PKCEExchangeRequest{
		Code:         code,
		CodeVerifier: verifier,
	})
	if err != nil {
		return errors.New("sign-in failed")
	}

	if err := s.applyTokens(tokens); err != nil {
		return errors.New("sign-in failed")
	}
	return nil
}

func (s *Session) applyTokens(tokens api.PKCETokenResponse) error {
	if tokens.RefreshToken == "" || tokens.AccessToken == "" {
		return fmt.Errorf("missing tokens")
	}
	if err := s.keys.Set(tokens.RefreshToken); err != nil {
		return err
	}
	s.mu.Lock()
	s.refreshToken = tokens.RefreshToken
	s.accessToken = tokens.AccessToken
	s.accessExpiry = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	s.signedIn = true
	s.mu.Unlock()
	s.api.SetAccessToken(tokens.AccessToken)
	s.notify(true, tokens.AccessToken)
	return nil
}

// AccessTokenForUI returns the current access token when signed in.
func (s *Session) AccessTokenForUI() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.signedIn {
		return ""
	}
	return s.accessToken
}

func (s *Session) SignedIn() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.signedIn
}

// Logout clears stored credentials.
func (s *Session) Logout() {
	_ = s.keys.Delete()
	s.mu.Lock()
	s.signedIn = false
	s.accessToken = ""
	s.accessExpiry = time.Time{}
	s.refreshToken = ""
	s.pending = nil
	s.mu.Unlock()
	s.api.SetAccessToken("")
	s.notify(false, "")
}

// Restore loads a refresh token from the keychain and refreshes the access token.
func (s *Session) Restore(ctx context.Context) error {
	refresh, err := s.keys.Get()
	if err != nil {
		return nil
	}
	tokens, err := s.api.PKCERefresh(ctx, refresh)
	if err != nil {
		_ = s.keys.Delete()
		return nil
	}
	if tokens.RefreshToken == "" {
		tokens.RefreshToken = refresh
	}
	if err := s.applyTokens(tokens); err != nil {
		_ = s.keys.Delete()
		return nil
	}
	return nil
}

// AccessToken returns a valid access token, refreshing when needed.
func (s *Session) AccessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	if !s.signedIn {
		s.mu.Unlock()
		return "", errors.New("not signed in")
	}
	needsRefresh := time.Until(s.accessExpiry) <= refreshSkew
	refreshTok := s.refreshToken
	token := s.accessToken
	s.mu.Unlock()

	if !needsRefresh {
		return token, nil
	}

	v, err, _ := s.refreshGroup.Do("refresh", func() (interface{}, error) {
		tokens, err := s.api.PKCERefresh(ctx, refreshTok)
		if err != nil {
			s.Logout()
			return "", err
		}
		if tokens.RefreshToken == "" {
			tokens.RefreshToken = refreshTok
		}
		if err := s.applyTokens(tokens); err != nil {
			s.Logout()
			return "", err
		}
		return tokens.AccessToken, nil
	})
	if err != nil {
		return "", errors.New("not signed in")
	}
	return v.(string), nil
}

// RunKeepalive refreshes tokens periodically until ctx is cancelled.
func (s *Session) RunKeepalive(ctx context.Context) {
	ticker := time.NewTicker(keepaliveInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !s.SignedIn() {
				continue
			}
			_, _ = s.AccessToken(ctx)
		}
	}
}

func (s *Session) notify(authenticated bool, accessToken string) {
	s.onChange(authenticated, accessToken)
}
