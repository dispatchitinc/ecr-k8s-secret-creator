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
		registries  []string
		success     bool
		name        string
		rendered    string
	}{
		{
			tokenOutput: &ecr.GetAuthorizationTokenOutput{
				AuthorizationData: []*ecr.AuthorizationData{
					&ecr.AuthorizationData{
						ProxyEndpoint:      aws.String("00000000.dkr.ecr.us-east-2.amazonaws.com"),
						AuthorizationToken: aws.String("xxx"),
					},
				}},
			registries: []string{},
			success:    true,
			name:       "with valid ecr token output",
			rendered:   "{\"auths\":{\"00000000.dkr.ecr.us-east-2.amazonaws.com\":{\"auth\":\"xxx\"}}}",
		},
		{
			tokenOutput: &ecr.GetAuthorizationTokenOutput{
				AuthorizationData: []*ecr.AuthorizationData{
					&ecr.AuthorizationData{
						ProxyEndpoint:      aws.String("001.dkr.ecr.us-east-2.amazonaws.com"),
						AuthorizationToken: aws.String("xxx"),
					},
				}},
			registries: []string{"002.dkr.ecr.us-east-2.amazonaws.com"},
			success:    true,
			name:       "with multiple registries in file",
			rendered:   "{\"auths\":{\"001.dkr.ecr.us-east-2.amazonaws.com\":{\"auth\":\"xxx\"},\"002.dkr.ecr.us-east-2.amazonaws.com\":{\"auth\":\"xxx\"}}}",
		},
		{
			tokenOutput: &ecr.GetAuthorizationTokenOutput{},
			registries:  []string{},
			success:     false,
			name:        "empty ecr token output",
			rendered:    "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			output, err := RenderDockerConfig(tc.tokenOutput, tc.registries)

			if tc.success {
				assert.NoError(t, err)
				assert.Equal(t, tc.rendered, string(output))
			} else {
				assert.Error(t, err)
			}
		})
	}
}
