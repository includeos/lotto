package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/mnordsletten/lotto/environment"
	"github.com/mnordsletten/lotto/mothership"
)

func mothershipFromConfig(filename string, env environment.Environment) (*mothership.Mothership, error) {
	m := &mothership.Mothership{}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return m, fmt.Errorf("error reading mothership config: %v", err)
	}
	if err = json.Unmarshal(file, m); err != nil {
		return m, fmt.Errorf("error decoding json: %v", err)
	}
	m, err = mothership.NewMothership(
		m.Host,
		m.Username,
		m.Password,
		m.Port,
		m.NoTLS,
		m.VerifyTLS,
		env)
	if err != nil {
		return m, fmt.Errorf("could not create new mothership: %v", err)
	}
	return m, nil
}
