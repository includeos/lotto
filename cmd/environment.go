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
			Host:         "",
			Org:          "",
			Username:     "",
			Password:     "",
			TemplateName: "",
			Catalog:      "",
			NetworkName:  "",
			SshRemote:    "lotto-client1",
			UplinkFile:   "vcloud-uplink.json",
		}
		return vcloud, nil
	case "openstack":
		return environment.NewOpenstack(), nil
	case "fusion":
		return environment.NewFusion(
			settings.Fusion.SshRemote,
			settings.Fusion.UplinkFile,
			settings.Fusion.VmspecPath), nil
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
