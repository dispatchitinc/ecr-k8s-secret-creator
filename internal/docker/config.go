package docker

import (
	"encoding/json"
	"errors"
)

type authConfig struct {
	Auth string `json:"auth"`
}

type dockerConfigAuthJson struct {
	Auths map[string]authConfig `json:"auths"`
}

func RenderDockerConfig(token string, registries []string) ([]byte, error) {
	if token == "" || len(registries) == 0 {
		return nil, errors.New("please provide a token and at least one registry")
	}

	doc := dockerConfigAuthJson{
		Auths: make(map[string]authConfig),
	}

	for _, registry := range registries {
		doc.Auths[registry] = authConfig{token}
	}

	rendered, err := json.Marshal(doc)
	if err != nil {
		return nil, errors.New("cannot render docker config auth json")
	}

	return rendered, nil
}
