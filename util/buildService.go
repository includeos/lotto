package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

func BuildServiceInDocker(servicePath, uplinkName, dockerContainerName string) error {
	// Delete all old resources
	buildFolder := path.Join(servicePath, "build")
	config := path.Join(servicePath, "config.json")
	for _, resource := range []string{buildFolder, config} {
		if err := os.RemoveAll(resource); err != nil {
			return fmt.Errorf("error removing resource %s: %v", resource, err)
		}
	}
	uplinkData, err := ioutil.ReadFile(uplinkName)
	if err != nil {
		return fmt.Errorf("error reading uplink %s: %v", uplinkName, err)
	}
	fullConfig := fmt.Sprintf("{\"uplink\":%s}", uplinkData)

	// Copy config.json to folder
	if err := ioutil.WriteFile(config, []byte(fullConfig), 0644); err != nil {
		return fmt.Errorf("error writing config to file: %v", err)
	}
	// Build using docker
	cur, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user info: %v", err)
	}

	cmdString := fmt.Sprintf("docker run -v %s:/service -u %s:%s %s", servicePath, cur.Uid, cur.Gid, dockerContainerName)
	if output, err := ExternalCommand(cmdString); err != nil {
		return fmt.Errorf("build failed: %s, error: %v", output, err)
	}
	return nil
}
