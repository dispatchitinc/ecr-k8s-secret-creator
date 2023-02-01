package authtoken

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/service/ecr"
)

type TokenProvider interface {
	GetAuthorizationToken(input *ecr.GetAuthorizationTokenInput) (*ecr.GetAuthorizationTokenOutput, error)
}

type AuthToken struct {
	provider TokenProvider

	Endpoint  string
	Token     string
	ExpiresAt time.Time
}

func NewAuthToken(provider TokenProvider) *AuthToken {
	return &AuthToken{
		provider: provider,
	}
}

func (t *AuthToken) Generate() error {
	tokenOutput, err := t.provider.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return err
	}

	if len(tokenOutput.AuthorizationData) < 1 {
		return errors.New("authorization data is empty")
	}

	t.Endpoint = *tokenOutput.AuthorizationData[0].ProxyEndpoint
	t.Token = *tokenOutput.AuthorizationData[0].AuthorizationToken
	t.ExpiresAt = *tokenOutput.AuthorizationData[0].ExpiresAt

	return nil
}
