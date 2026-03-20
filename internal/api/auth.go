package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type authContextKey string

const identityContextKey authContextKey = "api.identity"
const authModeContextKey authContextKey = "api.auth_mode"

type identity struct {
	Subject string
	Email   string
	Name    string
}

type kratosWhoamiResponse struct {
	Identity struct {
		ID     string `json:"id"`
		Traits struct {
			Email string `json:"email"`
			Name  struct {
				First string `json:"first"`
				Last  string `json:"last"`
			} `json:"name"`
		} `json:"traits"`
	} `json:"identity"`
}

func withAuth(next http.Handler, apiToken, kratosPublicURL string, allowInsecureAuth bool) http.Handler {
	client := &http.Client{Timeout: 10 * time.Second}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions || r.URL.Path == "/api/health" || r.URL.Path == "/api/ready" {
			next.ServeHTTP(w, r)
			return
		}

		if allowInsecureAuth && strings.TrimSpace(apiToken) == "" && strings.TrimSpace(kratosPublicURL) == "" {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authModeContextKey, "none")))
			return
		}

		if token := strings.TrimSpace(apiToken); token != "" {
			candidate := strings.TrimSpace(r.Header.Get("X-API-Token"))
			if candidate == "" {
				auth := strings.TrimSpace(r.Header.Get("Authorization"))
				if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
					candidate = strings.TrimSpace(auth[7:])
				}
			}
			if candidate == token {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authModeContextKey, "api_token")))
				return
			}
		}

		if publicURL := strings.TrimSpace(kratosPublicURL); publicURL != "" {
			ident, err := whoami(r.Context(), client, publicURL, r)
			if err == nil {
				ctx := context.WithValue(r.Context(), identityContextKey, ident)
				ctx = context.WithValue(ctx, authModeContextKey, "kratos")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	})
}

func whoami(ctx context.Context, client *http.Client, publicURL string, incoming *http.Request) (identity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(publicURL, "/")+"/sessions/whoami", nil)
	if err != nil {
		return identity{}, err
	}
	if cookie := strings.TrimSpace(incoming.Header.Get("Cookie")); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	if sessionToken := strings.TrimSpace(incoming.Header.Get("X-Session-Token")); sessionToken != "" {
		req.Header.Set("X-Session-Token", sessionToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return identity{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return identity{}, fmt.Errorf("kratos whoami status %d", resp.StatusCode)
	}

	var payload kratosWhoamiResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return identity{}, err
	}
	name := strings.TrimSpace(strings.Join([]string{
		strings.TrimSpace(payload.Identity.Traits.Name.First),
		strings.TrimSpace(payload.Identity.Traits.Name.Last),
	}, " "))
	return identity{
		Subject: payload.Identity.ID,
		Email:   strings.TrimSpace(payload.Identity.Traits.Email),
		Name:    strings.TrimSpace(name),
	}, nil
}

func identityFromContext(ctx context.Context) (identity, bool) {
	value, ok := ctx.Value(identityContextKey).(identity)
	return value, ok
}

func authModeFromContext(ctx context.Context) string {
	value, _ := ctx.Value(authModeContextKey).(string)
	if value == "" {
		return "none"
	}
	return value
}

func slugFromIdentity(ident identity) string {
	subject := strings.NewReplacer("-", "", "_", "").Replace(ident.Subject)
	if len(subject) > 12 {
		subject = subject[:12]
	}
	if subject == "" {
		subject = "user"
	}
	return "kratos-" + strings.ToLower(subject)
}

func displayNameFromIdentity(ident identity) string {
	if ident.Name != "" {
		return ident.Name
	}
	if ident.Email != "" {
		return ident.Email
	}
	return slugFromIdentity(ident)
}
