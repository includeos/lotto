package testFramework

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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

// ReadFromDisk takes a path and reads the testspec and returns a TestConfig
func ReadFromDisk(testPath string) (*TestConfig, error) {
	if err := verifyTestFiles(testPath); err != nil {
		return nil, err
	}
	test := &TestConfig{}
	test.testPath = testPath
	file, err := ioutil.ReadFile(path.Join(test.testPath, "testspec.json"))
	if err != nil {
		return test, fmt.Errorf("error reading test file %s: %v", test.testPath, err)
	}
	if err = json.Unmarshal(file, test); err != nil {
		return test, fmt.Errorf("error decoding json: %v", err)
	}
	test.NaclFile = path.Join(test.testPath, test.NaclFile)
	test.ClientCommandScript = path.Join(test.testPath, test.ClientCommandScript)
	return test, nil
}

// verifyTestFiles checks for the existence of selected files in the path supplied
func verifyTestFiles(testPath string) error {
	expectedFiles := []string{"testspec.json", "*.nacl", "*.sh"}

	for _, file := range expectedFiles {
		pathToCheck := path.Join(testPath, file)
		matches, err := filepath.Glob(pathToCheck)
		if err != nil {
			return fmt.Errorf("error looking for file: %s, %v", file, err)
		}
		if len(matches) < 1 {
			return fmt.Errorf("%s not found", file)
		}
	}

	return nil
}
