package testFramework

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

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

func ReadFromDisk(name string) (*TestConfig, error) {
	test := &TestConfig{}
	file, err := ioutil.ReadFile(name)
	if err != nil {
		return test, fmt.Errorf("error reading test file %s: %v", name, err)
	}
	if err = json.Unmarshal(file, test); err != nil {
		return test, fmt.Errorf("error decoding json: %v", err)
	}

	/*
		nacl, err := ioutil.ReadFile(test.NaclFile)
		if err != nil {
			return test, fmt.Errorf("error reading naclFile: %v", err)
		}
		test.Nacl = nacl
	*/

	return test, nil
}
