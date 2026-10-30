package identity

import (
	"context"
	"net/http"
	"os"
	"strings"

	"globalvideoplatform/services/api/internal/security"
	"globalvideoplatform/services/api/internal/store"
)

type contextKey string

const UserContextKey = contextKey("user")

type AuthMiddleware struct {
	Store          *store.MemoryStore
	JWTValidator   *security.Validator
	AllowDevTokens bool
}

func NewAuthMiddleware(s *store.MemoryStore) *AuthMiddleware {
	// Dev tokens are strictly gated: permitted only in test/development, prohibited in production
	env := strings.ToLower(os.Getenv("GVP_ENV"))
	allowDev := (env == "development" || env == "test") && os.Getenv("GVP_ALLOW_DEV_TOKENS") == "true"
	var validator *security.Validator
	issuer := os.Getenv("OIDC_ISSUER")
	audience := os.Getenv("OIDC_AUDIENCE")
	jwksURL := os.Getenv("OIDC_JWKS_URL")
	if issuer != "" && audience != "" && jwksURL != "" {
		validator = security.NewValidator(issuer, audience, jwksURL)
	}

	return &AuthMiddleware{
		Store:          s,
		JWTValidator:   validator,
		AllowDevTokens: allowDev,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Unauthenticated visitor context
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Invalid authorization header format")
			return
		}

		token := parts[1]
		var user *store.User

		// 1. Production OIDC / RS256 JWKS validation
		if m.JWTValidator != nil {
			claims, err := m.JWTValidator.Validate(r.Context(), token)
			if err != nil {
				WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Invalid or expired JWT token")
				return
			}
			// Resolve user by subject ID or verified email
			if u, ok := m.Store.Users[claims.Subject]; ok {
				user = u
			} else {
				for _, u := range m.Store.Users {
					if claims.EmailVerified && u.Email == claims.Email && claims.Email != "" {
						user = u
						break
					}
				}
			}
		}

		// 2. Controlled test harness lookup (STRICTLY restricted to test/dev environment)
		if user == nil && m.AllowDevTokens {
			if token == "admin-token" {
				user = m.Store.Users["11111111-1111-1111-1111-111111111111"]
			} else if token == "creator-token" {
				user = m.Store.Users["22222222-2222-2222-2222-222222222222"]
			} else if token == "viewer-token" {
				user = m.Store.Users["33333333-3333-3333-3333-333333333333"]
			} else if token == "advertiser-token" {
				user = m.Store.Users["44444444-4444-4444-4444-444444444444"]
			} else {
				// Exact user ID match in test mode
				if u, ok := m.Store.Users[token]; ok {
					user = u
				}
			}
		}

		// 3. If token was provided but could not be validated or user not found -> 401 Unauthorized
		if user == nil {
			WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Invalid or expired credentials")
			return
		}
		if !user.IsActive {
			WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Account is inactive")
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserContextKey).(*store.User)
		if !ok || user == nil {
			WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContextKey).(*store.User)
			if !ok || user == nil {
				WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
				return
			}

			hasRole := false
			for _, rName := range roles {
				if user.Role == rName || user.Role == "super_admin" {
					hasRole = true
					break
				}
			}

			if !hasRole {
				WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Access denied: insufficient role permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUser(ctx context.Context) *store.User {
	if u, ok := ctx.Value(UserContextKey).(*store.User); ok {
		return u
	}
	return nil
}

func GetUserFromContext(ctx context.Context) *store.User {
	return GetUser(ctx)
}
