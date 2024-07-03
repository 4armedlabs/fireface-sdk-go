package fireface

import (
	"context"

	"github.com/4armedlabs/fireface-sdk-go/auth"
)

const Version = "GitVersion"

type Config struct {
	// The URL of the Fireface server
	ServerURL string `json:"serverURL"`
}

type Option func(*App)

type App struct {
	serverURL string
	opts      []Option
}

func NewApp(ctx context.Context, config *Config, opts ...Option) (*App, error) {
	if config == nil {
		var err error
		config, err = getConfigDefaults()
		if err != nil {
			return nil, err
		}
	}

	app := &App{
		serverURL: config.ServerURL,
		opts:      opts,
	}

	for _, opt := range opts {
		opt(app)
	}

	return app, nil
}

func (a *App) Auth(ctx context.Context) (*auth.Client, error) {
	return nil, nil
}

func getConfigDefaults() (*Config, error) {
	return &Config{
		ServerURL: "https://fireface.4armedlabs.run",
	}, nil
}
