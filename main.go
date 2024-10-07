package main

import (
	"cmp"
	"context"
	"log/slog"
	"os"
)

type TokenSource interface {
	GetToken(ctx context.Context) (string, error)
}

type TokenWriter interface {
	Write(ctx context.Context, data []byte) error
}

type App struct {
	source TokenSource
	dest   TokenWriter
}

func main() {
	ctx := context.TODO()
	app, err := buildApp()
	if err != nil {
		slog.Error("could not build app", "err", err)
	}

	token, err := app.source.GetToken(ctx)
	if err != nil {
		slog.Error("could not get token", "err", err)
		os.Exit(1)
	}
	slog.Info("Token received")

	if err := app.dest.Write(ctx, []byte(token)); err != nil {
		slog.Error("could not write token", "err", err)
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
