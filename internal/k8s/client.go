package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ApplySecret(client kubernetes.Interface, content []byte, secretName, namespace string, kind v1.SecretType) error {
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
		},
		Data: data,
		Type: kind,
	}

	log.Info().Str("namespace", namespace).Str("name", secretName).Msg("creating secret")

	secretClient := client.CoreV1().Secrets(namespace)
	result, err := secretClient.Update(context.Background(), secret, metav1.UpdateOptions{})
	actionTaken := "updated"

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			result, err = secretClient.Create(context.Background(), secret, metav1.CreateOptions{})
			if err != nil {
				return err
			}

			actionTaken = "created"
		} else {
			return err
		}
	}

	log.Info().
		Str("secret", result.GetObjectMeta().GetName()).
		Msg(fmt.Sprintf("%s secret", actionTaken))

	return nil
}
