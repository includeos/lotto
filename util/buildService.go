package util

import (
	"fmt"
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

	// Copy config.json to folder
	if err := os.Link(uplinkName, config); err != nil {
		return fmt.Errorf("error linking config.json: %v", err)
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
