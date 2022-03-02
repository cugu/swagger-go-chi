package cli

import (
	"context"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc"
	"github.com/go-chi/chi"
	"golang.org/x/oauth2"

	"github.com/cugu/swagger-go-chi/testdata/customarray/generated/api"
	"github.com/cugu/swagger-go-chi/testdata/customarray/generated/auth"
)

type CLI struct {
	Debug bool `env:"DEBUG" default:"false"`
	Dev   bool `env:"DEV" default:"false"`

	OIDCURL           string   `name:"oidc-url"            env:"OIDC_URL"            required:""`
	OIDCIssuer        string   `name:"oidc-issuer"         env:"OIDC_ISSUER"         required:""`
	OIDCRedirectURL   string   `name:"oidc-redirect-url"   env:"OIDC_REDIRECT_URL"   required:""`
	OIDCClientID      string   `name:"oidc-client-id"      env:"OIDC_CLIENT_ID"      required:""`
	OIDCClientSecret  string   `name:"oidc-client-secret"  env:"OIDC_CLIENT_SECRET"  required:""`
	OIDCScopes        []string `name:"oidc-scopes"         env:"OIDC_SCOPES"                                      help:"Additional scopes, ['oidc', 'profile', 'email'] are always added." placeholder:"customscopes"`
	OIDCClaimUsername string   `name:"oidc-claim-username" env:"OIDC_CLAIM_USERNAME" default:"preferred_username" help:"username field in the OIDC claim"`
	OIDCClaimEmail    string   `name:"oidc-claim-email"    env:"OIDC_CLAIM_EMAIL"    default:"email"              help:"email field in the OIDC claim"`
	OIDCClaimName     string   `name:"oidc-claim-name"     env:"OIDC_CLAIM_NAME"     default:"name"               help:"name field in the OIDC claim"`
	AuthGroups        []string `env:"AUTH_GROUPS"`
	AuthDisabled      bool     `env:"AUTH_DISABLED"`
}

func NewServer(config CLI, s api.Service, fsys fs.FS) (chi.Router, error) {
	if config.Debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(io.Discard)
	}

	server := chi.NewRouter()

	var middlewares []func(next http.Handler) http.Handler
	if !config.AuthDisabled {
		// OIDC connection
		provider, err := oidc.NewProvider(context.Background(), config.OIDCIssuer)
		if err != nil {
			return nil, err
		}
		oauth2Config := oauth2.Config{
			ClientID:     config.OIDCClientID,
			ClientSecret: config.OIDCClientSecret,
			RedirectURL:  config.OIDCRedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		}
		verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

		middlewares = append(middlewares,
			auth.Required(config.OIDCURL, oauth2Config, verifier),
			auth.Group(config.AuthGroups...),
		)
		server.Get("/callback", auth.Callback(oauth2Config, verifier))
	}

	// server
	apiEndpoint := api.NewServer(s, api.IgnoreRoles, middlewares...)

	server.Mount("/api", apiEndpoint)

	staticHandler := api.Static(fsys)
	if config.Dev {
		log.Println("Use proxy")
		staticHandler = api.Proxy("http://localhost:8080")
	}
	server.Get("/manifest.json", staticHandler)
	server.NotFound(staticHandler)
	return server, nil
}
