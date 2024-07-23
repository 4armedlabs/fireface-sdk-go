package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/4armedlabs/fireface-sdk-go/api"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type client struct {
	secretKey string
	apiClient *api.Client
	baseURL   string
	jwksURL   string
	keySet    jwk.Set
	logger    *slog.Logger
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
	Logger    *slog.Logger
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

	httpClient := newHTTPClient()

	apiClient, err := api.NewClient(config.BaseURL, api.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	baseClient.apiClient = apiClient

	jwksCache := jwk.NewCache(ctx)
	jwksCache.Register(baseClient.jwksURL, jwk.WithMinRefreshInterval(5*time.Minute))

	// Refresh the cache immediately to populate
	keySet, err := jwksCache.Refresh(ctx, baseClient.jwksURL)
	if err != nil {
		return nil, err
	}

	baseClient.keySet = keySet
	baseClient.logger = config.Logger

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

	c.logger.Debug("verifying ID token", "idToken", idToken, "jwksURL", c.jwksURL)
	parsedToken, err := jwt.ParseString(idToken, jwt.WithKeySet(c.keySet))
	if err != nil {
		if c.keySet.Len() == 0 {
			c.logger.Error("key set is empty", "jwksURL", c.jwksURL)
		}

		var kid string
		k, ok := c.keySet.Key(0)
		if ok {
			kid = k.KeyID()
		}
		c.logger.Error("failed to parse ID token", "error", err, "keySetLength", c.keySet.Len(), "kid", kid)
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

type UserToUpdate struct {
	ID       string
	Email    string
	Password *string
}

type User struct {
	ID        string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *Client) UpdateUser(ctx context.Context, userToUpdate *UserToUpdate) (*User, error) {
	resp, err := c.apiClient.PutAuthUsersId(ctx, userToUpdate.ID, api.PutAuthUsersIdJSONRequestBody{
		Password: userToUpdate.Password,
	})
	if err != nil {
		return nil, err
	}

	user := api.User{}
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        *user.Id,
		Email:     user.Email,
		CreatedAt: time.Unix(user.CreatedAt, 0),
		UpdatedAt: time.Unix(user.UpdatedAt, 0),
	}, nil
}

type HttpClientOption func(*http.Client)

func newHTTPClient(opts ...HttpClientOption) *http.Client {
	hc := &http.Client{}

	for _, opt := range opts {
		opt(hc)
	}

	return hc
}
