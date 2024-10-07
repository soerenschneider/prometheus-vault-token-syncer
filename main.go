package main

import (
	"cmp"
	"context"
	"log/slog"
	"os"
	"time"
)

type TokenSource interface {
	Receive(ctx context.Context) (string, error)
	Cleanup(ctx context.Context) error
}

type TokenWriter interface {
	Write(ctx context.Context, data []byte) error
}

type App struct {
	source TokenSource
	dest   TokenWriter
}

func main() {
	app, err := buildApp()
	if err != nil {
		slog.Error("could not build app", "err", err)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	token, err := app.source.Receive(ctx)
	if err != nil {
		slog.Error("could not get token", "err", err)
		os.Exit(1)
	}
	slog.Info("Token received")

	if err := app.dest.Write(ctx, []byte(token)); err != nil {
		slog.Error("could not write token, trying to cleanup", "err", err)
		if err := app.source.Cleanup(ctx); err != nil {
			slog.Error("error while cleaning up token", "err", err)
		}
		os.Exit(1)
	}
	slog.Info("Wrote received token to configured storage")
}

func buildApp() (*App, error) {
	role := os.Getenv("VAULT_ROLE")
	mount := getEnvOrDefault("VAULT_MOUNT_PATH", "kubernetes")

	tokenSource, err := NewVaultTokenSource(role, mount)
	if err != nil {
		return nil, err
	}

	secretName := getEnvOrDefault("SECRET_NAME", "prometheus-vault-token")
	secretKey := getEnvOrDefault("SECRET_KEY", "vault-token")
	secretNamespace := getEnvOrDefault("SECRET_NAMESPACE", "default")

	tokenWriter, err := NewKubeTokenWriter(secretName, secretKey, secretNamespace)
	if err != nil {
		return nil, err
	}

	return &App{
		source: tokenSource,
		dest:   tokenWriter,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	return cmp.Or(os.Getenv(key), defaultValue)
}
