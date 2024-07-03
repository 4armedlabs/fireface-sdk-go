package auth

import (
	"context"
	"net/http"
)

type Client struct {
	secretKey  string
	httpClient *http.Client
}

type DecodedIdToken struct {
	Email         *string `json:"email,omitempty"`
	EmailVerified *bool   `json:"email_verified,omitempty"`
	Exp           *int64  `json:"exp,omitempty"`
	Iat           *int64  `json:"iat,omitempty"`
	Iss           *string `json:"iss,omitempty"`
	Sub           *string `json:"sub,omitempty"`
}

func (c *Client) VerifyIDToken(ctx context.Context, idToken string) (DecodedIdToken, error) {
	return DecodedIdToken{}, nil
}
