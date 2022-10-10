package docker

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go/service/ecr"
)

type authConfig struct {
	Auth string `json:"auth"`
}

type dockerConfigAuthJson struct {
	Auths map[string]authConfig `json:"auths"`
}

func RenderDockerConfig(ecrToken *ecr.GetAuthorizationTokenOutput, registries []string) ([]byte, error) {
	if len(ecrToken.AuthorizationData) < 1 {
		return nil, errors.New("Authorization data is empty")
	}

	registries = append(registries, *ecrToken.AuthorizationData[0].ProxyEndpoint)
	token := *ecrToken.AuthorizationData[0].AuthorizationToken

	doc := dockerConfigAuthJson{
		Auths: make(map[string]authConfig),
	}

	for _, registry := range registries {
		doc.Auths[registry] = authConfig{token}
	}

	rendered, err := json.Marshal(doc)
	if err != nil {
		return nil, errors.New("Cannot render docker config auth json")
	}

	return rendered, nil
}
