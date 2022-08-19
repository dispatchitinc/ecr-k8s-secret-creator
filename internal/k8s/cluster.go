package k8s

import (
	"io/ioutil"

	"github.com/rs/zerolog/log"
)

func GetNamespace() string {
	var namespace string

	output, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Error().Err(err).Msg("could not load current namespace")
		namespace = "default"
	} else {
		namespace = string(output)
	}

	return namespace
}
