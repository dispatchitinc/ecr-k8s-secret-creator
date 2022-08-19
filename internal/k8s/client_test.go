package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestApplySecret(t *testing.T) {
	err := ApplySecret(
		testclient.NewSimpleClientset(),
		[]byte("test"),
		"docker-secret",
		"test-namespace",
		v1.SecretTypeOpaque,
	)

	assert.NoError(t, err)
}
