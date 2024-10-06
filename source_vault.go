package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"
)

type VaultTokenSource struct {
	client *api.Client

	role      string
	mountPath string
}

func NewVaultTokenSource(role, mount string) (*VaultTokenSource, error) {
	config := api.DefaultConfig()

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	return &VaultTokenSource{
		client:    client,
		role:      role,
		mountPath: mount,
	}, nil
}

func (s *VaultTokenSource) GetToken(ctx context.Context) (string, error) {
	k8sAuth, err := auth.NewKubernetesAuth(
		s.role,
		auth.WithMountPath(s.mountPath),
	)
	if err != nil {
		return "", fmt.Errorf("unable to initialize Kubernetes auth method: %w", err)
	}

	secret, err := s.client.Auth().Login(ctx, k8sAuth)
	if err != nil {
		return "", fmt.Errorf("failed to login to Vault: %w", err)
	}

	return secret.Auth.ClientToken, nil
}
