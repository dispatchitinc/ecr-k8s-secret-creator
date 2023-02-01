package k8s

import (
	"context"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ApplySecret(client kubernetes.Interface, content []byte, secretName, namespace string, kind v1.SecretType) (*v1.Secret, error) {
	var data map[string][]byte

	switch kind {
	case v1.SecretTypeDockerConfigJson:
		data = map[string][]byte{
			string(v1.DockerConfigJsonKey): content,
		}
	case v1.SecretTypeDockercfg:
		data = map[string][]byte{
			string(v1.DockerConfigKey): content,
		}
	default:
		data = map[string][]byte{
			"config.json": content,
		}
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			Labels: map[string]string{
				"k8s.dispatchit.com/ecr-k8s-secret-creator": "1",
			},
		},
		Data: data,
		Type: kind,
	}

	secretClient := client.CoreV1().Secrets(namespace)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	secret, err := secretClient.Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			return secretClient.Create(ctx, secret, metav1.CreateOptions{})
		default:
			return nil, err
		}
	}

	return secret, nil
}
