package authtoken_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/dispatchitinc/ecr-k8s-secret-creator/internal/authtoken"
	"github.com/stretchr/testify/assert"
)

type mockProvider struct {
	expires time.Time
}

func (s mockProvider) GetAuthorizationToken(_ *ecr.GetAuthorizationTokenInput) (*ecr.GetAuthorizationTokenOutput, error) {
	return &ecr.GetAuthorizationTokenOutput{
		AuthorizationData: []*ecr.AuthorizationData{
			{
				AuthorizationToken: aws.String("weareauthorized"),
				ExpiresAt:          &s.expires,
				ProxyEndpoint:      aws.String("https://docker.dispatchit.com"),
			},
		},
	}, nil
}

func TestAuthTokenGenerate(t *testing.T) {
	provider := mockProvider{
		expires: time.Now().Add(time.Hour * 12),
	}

	token := authtoken.NewAuthToken(provider)
	err := token.Generate()
	assert.NoError(t, err)
	assert.Equal(t, token.Endpoint, "https://docker.dispatchit.com")
	assert.Equal(t, token.ExpiresAt, provider.expires)
	assert.Equal(t, token.Token, "weareauthorized")
}
