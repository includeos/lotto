package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/mnordsletten/lotto/environment"
)

func newEnvironment(targetEnv string, settings *environment.EnvSettings) (environment.Environment, error) {
	switch env := strings.ToLower(targetEnv); env {
	case "vcloud":
		vcloud := &environment.Vcloud{
			Host:         settings.Vcloud.Host,
			Org:          settings.Vcloud.Org,
			Username:     settings.Vcloud.Username,
			Password:     settings.Vcloud.Password,
			TemplateName: settings.Vcloud.TemplateName,
			Catalog:      settings.Vcloud.Catalog,
			NetworkName:  settings.Vcloud.NetworkName,
			UplinkFile:   settings.Vcloud.UplinkFile,
			Clients:      settings.Vcloud.Clients,
			Mothership:   settings.Vcloud.Mothership,
		}
		return vcloud, nil
	case "openstack":
		return environment.NewOpenstack(), nil
	case "fusion":
		return environment.NewFusion(
			settings.Fusion.Clients,
			settings.Fusion.UplinkFile,
			settings.Fusion.VmspecPath,
			settings.Fusion.Mothership), nil
	default:
		return nil, fmt.Errorf("%s not found", targetEnv)
	}
}

func envFromConfig(filename, targetEnv string) (environment.Environment, error) {
	settings := &environment.EnvSettings{}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading env-settings config: %v", err)
	}
	if err = json.Unmarshal(file, settings); err != nil {
		return nil, fmt.Errorf("error decoding json: %v", err)
	}
	return newEnvironment(targetEnv, settings)
}
