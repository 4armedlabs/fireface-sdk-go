package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type client struct {
	secretKey  string
	httpClient *http.Client
	baseURL    string
	jwksURL    string
	keySet     jwk.Set
}

type Client struct {
	*client
}

type AuthClientOption func(*client)

type AuthConfig struct {
	BaseURL   string
	SecretKey string
	Opts      []AuthClientOption
	Version   string
}

func NewClient(ctx context.Context, config *AuthConfig) (*Client, error) {
	baseClient := &client{}

	if config == nil {
		return nil, errors.New("config is nil")
	}

	if config.SecretKey == "" {
		return nil, errors.New("secret key is required")
	}

	baseClient.secretKey = config.SecretKey
	baseClient.baseURL = config.BaseURL
	baseClient.jwksURL = config.BaseURL + "/.well-known/jwks.json"

	baseClient.httpClient = newHTTPClient()

	jwksCache := jwk.NewCache(ctx)
	jwksCache.Register(baseClient.jwksURL, jwk.WithMinRefreshInterval(5*time.Minute))

	// Refresh the cache immediately to populate
	keySet, err := jwksCache.Refresh(ctx, baseClient.jwksURL)
	if err != nil {
		return nil, err
	}

	baseClient.keySet = keySet

	for _, opt := range config.Opts {
		opt(baseClient)
	}

	return &Client{
		client: baseClient,
	}, nil
}

type DecodedIdToken struct {
	Email         *string   `json:"email,omitempty"`
	EmailVerified *bool     `json:"email_verified,omitempty"`
	Exp           time.Time `json:"exp,omitempty"`
	Iat           time.Time `json:"iat,omitempty"`
	Iss           string    `json:"iss,omitempty"`
	Sub           string    `json:"sub,omitempty"`
}

func (c *Client) VerifyIDToken(ctx context.Context, idToken string) (DecodedIdToken, error) {
	var decodedIdToken DecodedIdToken

	parsedToken, err := jwt.Parse([]byte(idToken), jwt.WithKeySet(c.keySet))
	if err != nil {
		return decodedIdToken, err
	}

	if email, ok := parsedToken.Get("email"); ok {
		e := email.(string)
		decodedIdToken.Email = &e
	}

	if emailVerified, ok := parsedToken.Get("email_verified"); ok {
		ev := emailVerified.(bool)
		decodedIdToken.EmailVerified = &ev
	}

	decodedIdToken.Exp = parsedToken.Expiration()
	decodedIdToken.Iat = parsedToken.IssuedAt()
	decodedIdToken.Iss = parsedToken.Issuer()
	decodedIdToken.Sub = parsedToken.Subject()

	return decodedIdToken, nil
}

type HttpClientOption func(*http.Client)

func newHTTPClient(opts ...HttpClientOption) *http.Client {
	hc := &http.Client{}

	for _, opt := range opts {
		opt(hc)
	}

	return hc
}
