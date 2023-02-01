package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderDockerConfig(t *testing.T) {
	tt := []struct {
		token      string
		registries []string
		success    bool
		name       string
		rendered   string
	}{
		{
			token:      "xxx",
			registries: []string{"00000000.dkr.ecr.us-east-2.amazonaws.com"},
			success:    true,
			name:       "with valid ecr token output",
			rendered:   "{\"auths\":{\"00000000.dkr.ecr.us-east-2.amazonaws.com\":{\"auth\":\"xxx\"}}}",
		},
		{
			token:      "xxx",
			registries: []string{"001.dkr.ecr.us-east-2.amazonaws.com", "002.dkr.ecr.us-east-2.amazonaws.com"},
			success:    true,
			name:       "with multiple registries in file",
			rendered:   "{\"auths\":{\"001.dkr.ecr.us-east-2.amazonaws.com\":{\"auth\":\"xxx\"},\"002.dkr.ecr.us-east-2.amazonaws.com\":{\"auth\":\"xxx\"}}}",
		},
		{
			token:      "xxx",
			registries: []string{},
			success:    false,
			name:       "empty ecr token output",
			rendered:   "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			output, err := RenderDockerConfig(tc.token, tc.registries)

			if tc.success {
				assert.NoError(t, err)
				assert.Equal(t, tc.rendered, string(output))
			} else {
				assert.Error(t, err)
			}
		})
	}
}
