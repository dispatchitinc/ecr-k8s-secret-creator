package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

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
		output, err := parseSecretType(testCase.input)
		assert.NoError(t, err)

		if output != testCase.output {
			t.Errorf("Expected input %s to return output secret type %s", testCase.input, testCase.output)
		}
	}

	_, err := parseSecretType("")
	assert.Error(t, err)
}
