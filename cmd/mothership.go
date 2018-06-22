package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/mnordsletten/lotto/environment"
	"github.com/mnordsletten/lotto/mothership"
)

func mothershipFromConfig(filename string, env environment.Environment) (*mothership.Mothership, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading mothership config: %v", err)
	}
	allMotherships := map[string]mothership.Mothership{}
	if err = json.Unmarshal(file, &allMotherships); err != nil {
		return nil, fmt.Errorf("error decoding json: %v", err)
	}
	moth := allMotherships[env.GetMothershipName()]

	m, err := mothership.NewMothership(
		moth.Host,
		moth.Username,
		moth.Password,
		moth.Binary,
		moth.Port,
		moth.NoTLS,
		moth.VerifyTLS,
		env)
	if err != nil {
		return m, fmt.Errorf("could not create new mothership: %v", err)
	}
	return m, nil
}
