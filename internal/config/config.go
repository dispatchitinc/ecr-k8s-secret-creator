package config

import (
	"fmt"

	"github.com/caarlos0/env"
	v1 "k8s.io/api/core/v1"
)

type Config struct {
	Interval         int      `env:"INTERVAL" envDefault:"10"`
	AwsRegion        string   `env:"AWS_REGION" envDefault:"us-east-2"`
	SecretName       string   `env:"SECRET_NAME" envDefault:"ecr-docker-secret"`
	TargetNamespaces []string `env:"TARGET_NAMESPACES" envDefault:"default"`
	TargetRegistries []string `env:"TARGET_REGISTRIES"`
	SecretType       string   `env:"SECRET_TYPE" envDefault:"Opaque"`
	SecretTypeName   v1.SecretType
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (*Config, error) {
	config := &Config{}

	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	config.SecretTypeName, err = parseSecretType(config.SecretType)
	if err != nil {
		return config, err
	}

	return config, nil
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
