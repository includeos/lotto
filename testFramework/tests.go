package testFramework

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/mnordsletten/lotto/environment"
	"github.com/mnordsletten/lotto/mothership"
	"github.com/sirupsen/logrus"
)

type TestConfig struct {
	ID              string `json:"id"`
	Nacl            string
	ClientCommand   string `json:"clientcommand"`
	TestEnvironment environment.Environment
	Mothership      *mothership.Mothership
	Starbase        *mothership.Starbase
}

func (t *TestConfig) SaveToDisk() error {
	// Create data dir if it does not exist
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		os.Mkdir("data", 0755)
	}

	// Save instance to file
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}
	filename := path.Join("data", t.ID+".json")
	if err := ioutil.WriteFile(filename, b, 0600); err != nil {
		return err
	}
	return nil
}

func (t *TestConfig) TestLoop(env environment.Environment) error {
	logrus.Debugf("Started testloop for test: %s", t.ID)

	// Loop and monitor output from test
	return nil
}
