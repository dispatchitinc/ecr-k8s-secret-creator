package main

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/dispatchitinc/ecr-k8s-secret-creator/internal/config"
	"github.com/dispatchitinc/ecr-k8s-secret-creator/internal/docker"
	"github.com/dispatchitinc/ecr-k8s-secret-creator/internal/k8s"
	"github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("initializing")

	// Load application configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	log.Info().
		Str("region", cfg.AwsRegion).
		Int("interval", cfg.Interval).
		Str("secretName", cfg.SecretName).
		Strs("targetNamespaces", cfg.TargetNamespaces).
		Str("secretType", cfg.SecretType).
		Msg("config loaded")

	sess := session.Must(session.NewSession(&aws.Config{Region: &cfg.AwsRegion}))
	svc := ecr.New(sess)

	// The channel that detects secret update requests
	refresh := make(chan bool)
	go func() { refresh <- true }()

	// The timer that refreshes the secrets before it expires
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Info().Msg("tick")
			go func() { refresh <- true }()
		case <-refresh:
			log.Info().Msg("refreshing secrets")

			refreshSecrets(svc, &cfg)

			log.Info().Msg("finished refreshing")
		}
	}
}

func refreshSecrets(svc *ecr.ECR, cfg *config.Config) {
	// Get the ECR authorization token from AWS
	tokenInput := &ecr.GetAuthorizationTokenInput{}

	token, err := svc.GetAuthorizationToken(tokenInput)
	if err != nil {
		log.Error().Err(err).Msg("could not get authorization token")
	}

	// Create the docker config.json in buffer
	dockerCfg, err := docker.RenderDockerConfig(token)
	if err != nil {
		log.Error().Err(err).Msg("could not create a docker config")
	}

	// Create the docker config.json as a kubernetes secret
	kconfig, err := rest.InClusterConfig()
	if err != nil {
		log.Error().Err(err).Msg("could not load docker config.json")
	}
	clientSet, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		log.Error().Err(err).Msg("could not initialize a new client set")
	}

	for _, ns := range cfg.TargetNamespaces {
		// Create or update the secret with the latest information
		k8s.ApplySecret(clientSet, dockerCfg, cfg.SecretName, ns, cfg.SecretTypeName)
		if err != nil {
			log.Error().Err(err).Msg("could not apply the docker secret")
		}
	}
}
