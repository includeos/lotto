package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/mnordsletten/lotto/environment"
)

// envFromConfig reads the config-environment.json file and does the following:
// 1. Figures out which environment you want to run. targetEnv refers to a name
// 2. Takes that name and finds out which type of environment it is (envType)
// 3. Decodes the envType to the Environment interface and returns to the user
func envFromConfig(filename, targetEnv string) (environment.Environment, error) {
	var env environment.Environment
	// All configs must include this field to declare which env type it is, e.g. vcloud, fusion
	type envType struct {
		EnvType string `json:"envType"`
	}

	// Read config file
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading env-settings config: %v", err)
	}
	// First extract the names of the different configs
	var configFileFull map[string]json.RawMessage
	if err = json.Unmarshal(file, &configFileFull); err != nil {
		return nil, fmt.Errorf("error decoding json: %v", err)
	}
	// Loop through the names, checking what type they are (envType), and then create the correct config
	for envName := range configFileFull {
		// configContent is the content under that particular name
		configContent := configFileFull[envName]
		var envT envType
		if err = json.Unmarshal(configContent, &envT); err != nil {
			return nil, fmt.Errorf("error decoding json: %v", err)
		}
		// Keep on going if this is not the env we are looking for
		if envName != targetEnv {
			continue
		}

		// create the correct configuration here
		switch envT.EnvType {
		case "vcloud":
			env = &environment.Vcloud{}
		case "fusion":
			env = &environment.Fusion{}
		case "openstack":
			return env, fmt.Errorf("Openstack not yet implemented")
		default:
			return env, fmt.Errorf("%s environment not found, or missing", envT.EnvType)
		}
		if err = json.Unmarshal(configContent, &env); err != nil {
			return nil, fmt.Errorf("error decoding json: %v", err)
		}
		// set the name of the config chosen
		env.SetName(envName)
		return env, nil
	}
	return env, fmt.Errorf("Environment %s not found", targetEnv)
}
