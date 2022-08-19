package docker

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/stretchr/testify/assert"
)

func TestRenderDockerConfig(t *testing.T) {
	tt := []struct {
		tokenOutput *ecr.GetAuthorizationTokenOutput
		success     bool
		name        string
	}{
		{
			tokenOutput: &ecr.GetAuthorizationTokenOutput{
				AuthorizationData: []*ecr.AuthorizationData{
					&ecr.AuthorizationData{
						ProxyEndpoint:      aws.String("xxx"),
						AuthorizationToken: aws.String("xxx"),
					},
				}},
			success: true,
			name:    "with valid ecr token output",
		},
		{
			tokenOutput: &ecr.GetAuthorizationTokenOutput{},
			success:     false,
			name:        "empty ecr token output",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, err := RenderDockerConfig(tc.tokenOutput)

			if tc.success {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
