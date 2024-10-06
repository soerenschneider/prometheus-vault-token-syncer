package main

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubeTokenWriter struct {
	namespace  string
	secretName string
	secretKey  string

	client *kubernetes.Clientset
}

func NewKubeTokenWriter(secretName, secretKey, namespace string) (*KubeTokenWriter, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &KubeTokenWriter{
		client:     clientset,
		secretName: secretName,
		secretKey:  secretKey,
		namespace:  namespace,
	}, nil
}

func (w *KubeTokenWriter) Write(ctx context.Context, data []byte) error {
	secretData := map[string][]byte{
		w.secretKey: data,
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: w.secretName,
		},
		Type: v1.SecretTypeOpaque,
		Data: secretData,
	}

	_, err := w.client.CoreV1().Secrets(w.namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			_, err = w.client.CoreV1().Secrets(w.namespace).Update(ctx, secret, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update secret: %v", err)
			}
		} else {
			return fmt.Errorf("failed to write secret: %v", err)
		}
	}

	return err
}
