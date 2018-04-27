package cmd

import (
	"fmt"
	"strings"

	"github.com/mnordsletten/lotto/environment"
)

func newEnvironment(targetEnv string, settings environment.EnvSettings) (environment.Environment, error) {
	switch env := strings.ToLower(targetEnv); env {
	case "vcloud":
		return environment.NewVcloud(settings), nil
	case "openstack":
		return environment.NewOpenstack(), nil
	default:
		return nil, fmt.Errorf("%s not found", targetEnv)
	}
}
