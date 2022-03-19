package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"

	"github.com/cugu/swagger-go-chi/testdata/customarray/generated/api"
)

type ContextKey string

const (
	stateSession            = "state"
	userSession             = "user"
	UserContext  ContextKey = "user"
)

func Required(oidcURL string, oauth2Config oauth2.Config, verifier *oidc.IDTokenVerifier) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader != "" {
				bearerAuth(oidcURL, authHeader, verifier)(next).ServeHTTP(w, r)

				return
			}
			sessionAuth(oauth2Config)(next).ServeHTTP(w, r)
		})
	}
}

type Claims struct {
	Acr               string   `json:"acr"`
	AllowedOrigins    []string `json:"allowed-origins"`
	AuthTime          int      `json:"auth_time"`
	Azp               string   `json:"azp"`
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	Exp               int      `json:"exp"`
	FamilyName        string   `json:"family_name"`
	GivenName         string   `json:"given_name"`
	Groups            []string `json:"groups"`
	Iat               int      `json:"iat"`
	Iss               string   `json:"iss"`
	Jti               string   `json:"jti"`
	Name              string   `json:"name"`
	PreferredUsername string   `json:"preferred_username"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess struct {
		App struct {
			Roles []string `json:"roles"`
		} `json:"app"`
	} `json:"resource_access"`
	Scope        string `json:"scope"`
	SessionState string `json:"session_state"`
	Sub          string `json:"sub"`
	Typ          string `json:"typ"`
}

func bearerAuth(oidcURL string, authHeader string, verifier *oidc.IDTokenVerifier) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(authHeader, "Bearer ") {
				api.JSONErrorStatus(w, http.StatusUnauthorized, errors.New("no bearer token"))
				return
			}

			authToken, err := verifier.Verify(r.Context(), authHeader[7:])
			if err != nil {
				api.JSONError(w, fmt.Errorf("could not verify bearer token: %v", err))

				return
			}

			claims := &Claims{}
			if err := authToken.Claims(claims); err != nil {
				api.JSONError(w, fmt.Errorf("failed to parse claims: %v", err))

				return
			}

			if claims.Iss != oidcURL {
				api.JSONError(w, fmt.Errorf("wrong issuer"))

				return
			}

			// set user session cookie
			b, _ := json.Marshal(claims)
			http.SetCookie(w, &http.Cookie{Name: userSession, Value: base64.StdEncoding.EncodeToString(b)})

			// set user context
			r = r.WithContext(context.WithValue(r.Context(), UserContext, claims))

			next.ServeHTTP(w, r)
		})
	}
}

func sessionAuth(oauth2Config oauth2.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCookie, err := r.Cookie(userSession)
			if err != nil || userCookie == nil {
				redirectToLogin(w, r, oauth2Config)

				return
			}

			b, err := base64.StdEncoding.DecodeString(userCookie.Value)
			if err != nil {
				api.JSONError(w, errors.New("could not decode session"))
				return
			}

			claims := &Claims{}
			if err := json.Unmarshal(b, claims); err != nil {
				api.JSONError(w, errors.New("claims not in session"))
				return
			}

			// set user context
			r = r.WithContext(context.WithValue(r.Context(), UserContext, claims))

			next.ServeHTTP(w, r)
		})
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request, oauth2Config oauth2.Config) {
	state, err := state()
	if err != nil {
		api.JSONError(w, fmt.Errorf("generating state failed"))

		return
	}

	http.SetCookie(w, &http.Cookie{Name: stateSession, Value: base64.StdEncoding.EncodeToString([]byte(state))})

	http.Redirect(w, r, oauth2Config.AuthCodeURL(state), http.StatusFound)
}

func Callback(oauth2Config oauth2.Config, verifier *oidc.IDTokenVerifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stateCookie, err := r.Cookie(stateSession)
		if err != nil {
			api.JSONError(w, fmt.Errorf("state missing"))

			return
		}

		b, err := base64.StdEncoding.DecodeString(stateCookie.Value)
		if err != nil {
			api.JSONError(w, fmt.Errorf("could not decode state"))

			return
		}

		if string(b) != r.URL.Query().Get("state") {
			api.JSONError(w, fmt.Errorf("state mismatch"))

			return
		}

		oauth2Token, err := oauth2Config.Exchange(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			api.JSONError(w, fmt.Errorf("oauth2 exchange failed"))

			return
		}

		// Extract the ID Token from OAuth2 token.
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			api.JSONError(w, fmt.Errorf("missing id token"))

			return
		}

		// Parse and verify ID Token payload.
		idToken, err := verifier.Verify(r.Context(), rawIDToken)
		if err != nil {
			api.JSONError(w, fmt.Errorf("token verification failed"))

			return
		}

		// Extract custom claims
		claims := &Claims{}
		if err := idToken.Claims(claims); err != nil {
			api.JSONError(w, fmt.Errorf("claim extraction failed"))

			return
		}

		// set user session cookie
		b, _ = json.Marshal(claims)
		http.SetCookie(w, &http.Cookie{Name: userSession, Value: base64.StdEncoding.EncodeToString(b)})

		// set user context
		r = r.WithContext(context.WithValue(r.Context(), UserContext, claims))

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func Group(allowedGroups ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContext).(*Claims)
			if ok {
				if !(contains(user.Groups, allowedGroups)) { // || contains([]string{"service"}, user.ResourceAccess.App.Roles)) {
					api.JSONErrorStatus(w, http.StatusUnauthorized, errors.New("group not allowed"))
					return
				}
			} else {
				api.JSONErrorStatus(w, http.StatusUnauthorized, errors.New("no user in context"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func state() (string, error) {
	rnd := make([]byte, 32)
	if _, err := rand.Read(rnd); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(rnd), nil
}

func contains(userGroups, allowedGroups []string) bool {
	for _, allowedGroup := range allowedGroups {
		for _, userGroup := range userGroups {
			if userGroup == allowedGroup {
				return true
			}
		}
	}

	return false
}
