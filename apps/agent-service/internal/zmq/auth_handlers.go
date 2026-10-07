package zmq

import (
	"context"
	"encoding/json"

	"github.com/nayan-bagale/skydock-agent/internal/auth"
	"github.com/nayan-bagale/skydock-agent/internal/config"
)

const (
	EventAuthStart    = "AUTH_START"
	EventAuthCallback = "AUTH_CALLBACK"
	EventAuthSession  = "AUTH_SESSION"
	EventAuthLogout   = "AUTH_LOGOUT"
	EventAuthChanged  = "auth:changed"
)

type authCallbackBody struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func registerAuthHandlers(s *Server, session *auth.Session) {
	s.On(EventAuthStart, handleAuthStart(session))
	s.On(EventAuthCallback, handleAuthCallback(session))
	s.On(EventAuthSession, handleAuthSession(session))
	s.On(EventAuthLogout, handleAuthLogout(session))
}

func handleAuthStart(session *auth.Session) Handler {
	return func(ctx context.Context, _ string, _ json.RawMessage) (any, error) {
		loginURL, err := session.Start(ctx)
		if err != nil {
			return map[string]interface{}{
				"ok":    false,
				"error": "sign-in failed",
			}, nil
		}
		return map[string]interface{}{
			"ok":  true,
			"url": loginURL,
		}, nil
	}
}

func handleAuthCallback(session *auth.Session) Handler {
	return func(ctx context.Context, _ string, data json.RawMessage) (any, error) {
		var body authCallbackBody
		if len(data) > 0 {
			if err := json.Unmarshal(data, &body); err != nil {
				return map[string]interface{}{
					"ok":    false,
					"error": "sign-in failed",
				}, nil
			}
		}
		err := session.Complete(ctx, body.Code, body.State)
		if err != nil {
			msg := err.Error()
			if msg != "sign-in expired" && msg != "sign-in failed" {
				msg = "sign-in failed"
			}
			return map[string]interface{}{
				"ok":    false,
				"error": msg,
			}, nil
		}
		return map[string]interface{}{
			"ok":            true,
			"authenticated": true,
		}, nil
	}
}

func handleAuthSession(session *auth.Session) Handler {
	return func(ctx context.Context, _ string, _ json.RawMessage) (any, error) {
		return map[string]interface{}{
			"ok":            true,
			"authenticated": session.SignedIn(),
			"accessToken":   session.AccessTokenForUI(),
			"apiBaseUrl":    config.Current().APIBaseURL,
		}, nil
	}
}

func handleAuthLogout(session *auth.Session) Handler {
	return func(ctx context.Context, _ string, _ json.RawMessage) (any, error) {
		session.Logout()
		return map[string]interface{}{
			"ok":            true,
			"authenticated": false,
		}, nil
	}
}

// AuthChangeEmitter notifies desktop peers when session state changes.
type AuthChangeEmitter func(authenticated bool, accessToken string)

func NewAuthChangeEmitter(server *Server) AuthChangeEmitter {
	return func(authenticated bool, accessToken string) {
		if server == nil {
			return
		}
		_ = server.EmitAll(EventAuthChanged, map[string]interface{}{
			"authenticated": authenticated,
			"accessToken":   accessToken,
			"apiBaseUrl":    config.Current().APIBaseURL,
		})
	}
}
