package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestLoadingTargetNamespaces(t *testing.T) {
	var testCases = []struct {
		env    string
		result []string
	}{
		{"", []string{}},
		{"default", []string{"default"}},
		{"default,dispatch", []string{"default", "dispatch"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.env, func(t *testing.T) {
			os.Setenv("TARGET_NAMESPACES", testCase.env)
			cfg, err := LoadConfig()
			assert.NoError(t, err)
			assert.ElementsMatch(t, testCase.result, cfg.TargetNamespaces)
		})
	}
}

func TestLoadingTargetRegistries(t *testing.T) {
	var testCases = []struct {
		env    string
		result []string
	}{
		{"", []string{}},
		{"00000000.dkr.ecr.us-east-2.amazonaws.com", []string{"00000000.dkr.ecr.us-east-2.amazonaws.com"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.env, func(t *testing.T) {
			os.Setenv("TARGET_REGISTRIES", testCase.env)
			cfg, err := LoadConfig()
			assert.NoError(t, err)
			assert.ElementsMatch(t, testCase.result, cfg.TargetRegistries)
		})
	}
}

func TestParseSecretType(t *testing.T) {
	var testCases = []struct {
		input  string
		output v1.SecretType
	}{
		{"Opaque", v1.SecretTypeOpaque},
		{"kubernetes.io/dockerconfigjson", v1.SecretTypeDockerConfigJson},
		{"kubernetes.io/dockercfg", v1.SecretTypeDockercfg},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			output, err := parseSecretType(testCase.input)
			assert.NoError(t, err)

			if output != testCase.output {
				t.Errorf("Expected input %s to return output secret type %s", testCase.input, testCase.output)
			}
		})
	}

	_, err := parseSecretType("")
	assert.Error(t, err)
}
