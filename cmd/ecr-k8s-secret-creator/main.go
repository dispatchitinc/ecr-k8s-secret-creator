package main

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/panbanda/ecr-k8s-secret-creator/internal/config"
	"github.com/panbanda/ecr-k8s-secret-creator/internal/docker"
	"github.com/panbanda/ecr-k8s-secret-creator/internal/k8s"
	"github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const appVersion = "0.1.0"

func main() {
	if os.Getenv("APP_ENV") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Info().Str("version", appVersion).Msg("initializing")

	// Load application configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	log.Info().
		Str("region", cfg.AwsRegion).
		Int("interval", cfg.Interval).
		Str("secretName", cfg.SecretName).
		Str("secretType", cfg.SecretType).
		Msg("loaded flags")

	sess := session.Must(session.NewSession(&aws.Config{Region: &cfg.AwsRegion}))
	svc := ecr.New(sess)

	for {
		// Get the ECR authorization token from AWS
		tokenInput := &ecr.GetAuthorizationTokenInput{}

		token, err := svc.GetAuthorizationToken(tokenInput)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get authorization token")
		}

		// Create the docker config.json in buffer
		dockerCfg, err := docker.RenderDockerConfig(token)
		if err != nil {
			log.Fatal().Err(err).Msg("could not create a docker config")
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

		// Create or update the secret with the latest information
		k8s.ApplySecret(clientSet, dockerCfg, cfg.SecretName, k8s.GetNamespace(), cfg.SecretTypeName)
		if err != nil {
			log.Fatal().Err(err).Msg("could not apply the docker secret")
		}

		// Sleep until the next refresh cycle
		time.Sleep(time.Duration(cfg.Interval) * time.Second)
	}
}
