package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/dispatchitinc/ecr-k8s-secret-creator/internal/authtoken"
	"github.com/dispatchitinc/ecr-k8s-secret-creator/internal/config"
	"github.com/dispatchitinc/ecr-k8s-secret-creator/internal/docker"
	"github.com/dispatchitinc/ecr-k8s-secret-creator/internal/k8s"
	"github.com/dispatchitinc/ecr-k8s-secret-creator/internal/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	log    *zap.SugaredLogger
	token  *authtoken.AuthToken
	cfg    *config.Config
	klient *kubernetes.Clientset
)

func main() {
	var err error

	log, err = logger.New("ecr-secret-creator")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("initializing")

	// Load application configuration
	cfg, err = config.LoadConfig()
	if err != nil {
		log.Fatalw("failed to load configuration", "error", err)
	}

	log.Infow(
		"loaded configuration",
		"region", cfg.AwsRegion,
		"secretName", cfg.SecretName,
		"targetNamespaces", cfg.TargetNamespaces,
		"secretType", cfg.SecretType,
	)

	sess := session.Must(session.NewSession(&aws.Config{Region: &cfg.AwsRegion}))
	svc := ecr.New(sess)
	token = authtoken.NewAuthToken(svc)

	kconfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalw("cannot get cluster config", "error", err)
	}

	klient, err = kubernetes.NewForConfig(kconfig)
	if err != nil {
		log.Fatalw("cannot init kubernetes client from config", "error", err)
	}

	// The channel that detects secret update requests
	refresh := make(chan bool)
	go func() { refresh <- true }()

	// The timer that refreshes the secrets before it expires
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Info("tick")
			go func() { refresh <- true }()
		case <-refresh:
			// Refresh the token if it is within 3 hours
			future := time.Now().Add(time.Hour * 3)
			if token.ExpiresAt.Before(future) {
				err := token.Generate()
				if err != nil {
					log.Errorw("could not refresh token", "error", err)
					continue
				}

				log.Info("generated new token")
			}

			err := refreshSecrets()
			if err != nil {
				log.Errorw("could not refresh secrets", "error", err)
			}
		}
	}
}

func refreshSecrets() error {
	log.Info("refreshing secrets")

	labelList := map[string]string{
		"k8s.dispatchit.com/ecr-k8s-secret-creator": "1",
	}
	secretList, err := klient.CoreV1().Secrets("").List(context.Background(), v1.ListOptions{
		LabelSelector: labels.Set(labelList).String(),
	})
	if err != nil {
		return err
	}

	existingNS := map[string]bool{}
	managed := managedNamespaces()
	for _, secret := range secretList.Items {
		// Previously created secrets that are no longer in a managed
		// namespace because helm chart changed the managed namespaces
		_, isManaged := managed[secret.Namespace]
		if !isManaged {
			log.Info("namespace not managed, deleting secret", "name", secret.Name, "namespace", secret.Namespace)
			deleteSecret(secret)
			continue
		}

		// Secrets that were created under an old name configuration
		if secret.Name != cfg.SecretName {
			log.Info("mismatched name, deleting secret", "name", secret.Name, "namespace", secret.Namespace)
			deleteSecret(secret)
			continue
		}

		existingNS[secret.Namespace] = true
	}

	// Create the docker config.json in buffer
	registries := cfg.TargetRegistries
	registries = append(registries, token.Endpoint)
	dockerCfg, err := docker.RenderDockerConfig(token.Token, registries)
	if err != nil {
		log.Errorw("could not create a docker config", "registries", registries)
	}

	for _, ns := range cfg.TargetNamespaces {
		_, exists := existingNS[ns]
		if exists {
			log.Infof("secret already exists in %s, skipping", ns)
			continue
		}

		// Create or update the secret with the latest information
		secret, err := k8s.ApplySecret(klient, dockerCfg, cfg.SecretName, ns, cfg.SecretTypeName)
		switch {
		case err != nil:
			log.Errorw("could not apply the docker secret", "error", err)
		default:
			log.Info("created secret", "name", secret.Name, "namespace", secret.Namespace)
		}
	}

	log.Info("finished refreshing")

	return nil
}

func deleteSecret(secret corev1.Secret) error {
	return klient.CoreV1().Secrets(secret.Namespace).Delete(context.Background(), secret.Name, v1.DeleteOptions{})
}

func managedNamespaces() map[string]bool {
	nsmap := map[string]bool{}
	for _, ns := range cfg.TargetNamespaces {
		nsmap[ns] = true
	}

	return nsmap
}
