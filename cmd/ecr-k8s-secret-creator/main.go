package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	log "github.com/rs/zerolog/log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const appVersion = "0.1.0"

const cfgTemplate = `{
  "auths": {
    "{{ .registry }}": {
      "auth": "{{ .token }}"
    }
  }
}`

func main() {
	log.Info().Str("version", appVersion).Msg("initializing")

	region := flag.String("region", "us-east-1", "The aws region")
	interval := flag.Int("interval", 1200, "Refresh interval in seconds")
	profile := flag.String("profile", "", "The AWS Account profile")
	secretName := flag.String("secretName", "ecr-login-password", "The name of the secret")
	secretTypeName := flag.String("secretType", "Opaque", fmt.Sprintf("The secret type, available options: (%s|%s|%s)",
		v1.SecretTypeOpaque, v1.SecretTypeDockerConfigJson, v1.SecretTypeDockercfg))

	flag.Parse()

	log.Info().
		Str("region", *region).
		Str("profile", *profile).
		Int("interval", *interval).
		Str("secretName", *secretName).
		Str("secretType", *secretTypeName).
		Msg("loaded flags")

	secretType, err := parseSecretType(*secretTypeName)
	if err != nil {
		log.Fatal().Err(err).Msg("could not parse the secret type")
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: region,
	}))
	svc := ecr.New(sess)

	for {
		// Get the ECR authorization token from AWS
		tokenInput := &ecr.GetAuthorizationTokenInput{}
		if *profile != "" {
			tokenInput.RegistryIds = []*string{profile}
		}

		token, err := svc.GetAuthorizationToken(tokenInput)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get authorization token")
		}

		// Create the docker config.json in buffer
		dockerCfg, err := createDockerCfg(token)

		// Get current namespace
		namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			log.Fatal().Err(err).Msg("could not load current namespace")
		}

		// Create the docker config.json as a kubernetes secret
		kconfig, err := rest.InClusterConfig()
		if err != nil {
			log.Fatal().Err(err).Msg("could not load docker config.json")
		}
		clientSet, err := kubernetes.NewForConfig(kconfig)
		if err != nil {
			log.Fatal().Err(err).Msg("could not initialize a new client set")
		}
		kclient := &kubernetesAPI{client: clientSet}
		err = kclient.applyDockerCfgSecret(dockerCfg, *secretName, string(namespace), secretType)
		if err != nil {
			log.Fatal().Err(err).Msg("could not apply the docker secret")
		}

		// Sleep until the next refresh cycle
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}

func createDockerCfg(ecrToken *ecr.GetAuthorizationTokenOutput) ([]byte, error) {
	if len(ecrToken.AuthorizationData) < 1 {
		return nil, errors.New("authorization data should have at least 1 auth data")
	}

	cfgData := map[string]string{}
	cfgData["registry"] = *ecrToken.AuthorizationData[0].ProxyEndpoint
	cfgData["token"] = *ecrToken.AuthorizationData[0].AuthorizationToken

	// Put the config template output in a buffer
	t := template.Must(template.New("").Parse(cfgTemplate))
	cfgBuffer := bytes.NewBufferString("")

	if err := t.Execute(cfgBuffer, cfgData); err != nil {
		return nil, err
	}

	cfgInByte, err := ioutil.ReadAll(cfgBuffer)
	if err != nil {
		return nil, err
	}

	return cfgInByte, nil
}

type kubernetesAPI struct {
	client kubernetes.Interface
}

func (k *kubernetesAPI) applyDockerCfgSecret(cfg []byte, secretName, namespace string, kind v1.SecretType) error {
	var data map[string][]byte
	switch kind {
	case v1.SecretTypeDockerConfigJson:
		data = map[string][]byte{
			string(v1.DockerConfigJsonKey): cfg,
		}
	case v1.SecretTypeDockercfg:
		data = map[string][]byte{
			string(v1.DockerConfigKey): cfg,
		}
	default:
		data = map[string][]byte{
			"config.json": cfg,
		}
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: data,
		Type: kind,
	}

	log.Info().Msg("creating secret")

	secretClient := k.client.CoreV1().Secrets(namespace)
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

func parseSecretType(s string) (v1.SecretType, error) {
	switch v1.SecretType(s) {
	case v1.SecretTypeOpaque:
		return v1.SecretTypeOpaque, nil
	case v1.SecretTypeDockerConfigJson:
		return v1.SecretTypeDockerConfigJson, nil
	case v1.SecretTypeDockercfg:
		return v1.SecretTypeDockercfg, nil
	default:
		return "", fmt.Errorf("unmatched secret type: %s", s)
	}
}
