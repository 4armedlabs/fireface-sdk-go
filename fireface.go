package fireface

import (
	"context"
	"log/slog"
	"os"

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
	logger    *slog.Logger
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

	if app.logger == nil {
		logOptions := &slog.HandlerOptions{}
		if os.Getenv("FIREFACE_DEBUG") == "true" {
			logOptions.Level = slog.LevelDebug
			logOptions.AddSource = true
		}

		logger := slog.New(slog.NewJSONHandler(os.Stdout, logOptions))
		logger.With("service", "fireface-sdk-go")

		app.logger = logger
	}

	return app, nil
}

func (a *App) Auth(ctx context.Context) (*auth.Client, error) {
	return auth.NewClient(ctx, &auth.AuthConfig{
		BaseURL: a.serverURL,
		Version: Version,
		Logger:  a.logger,
	})
}

func getConfigDefaults() (*Config, error) {
	return &Config{
		ServerURL: "https://fireface.4armedlabs.run",
	}, nil
}
