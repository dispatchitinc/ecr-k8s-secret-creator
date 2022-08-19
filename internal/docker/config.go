package docker

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ecr"
)

const dockerConfigTemplate = `{"auths": {"%s": {"auth": "%s"}}}`

func RenderDockerConfig(ecrToken *ecr.GetAuthorizationTokenOutput) ([]byte, error) {
	if len(ecrToken.AuthorizationData) < 1 {
		return nil, errors.New("Authorization data is empty")
	}

	endpoint := *ecrToken.AuthorizationData[0].ProxyEndpoint
	token := *ecrToken.AuthorizationData[0].AuthorizationToken
	rendered := fmt.Sprintf(dockerConfigTemplate, endpoint, token)

	return []byte(rendered), nil
}
