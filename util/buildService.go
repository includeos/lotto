package util

import "fmt"

func BuildServiceInDocker(path string) error {
	cmdString := fmt.Sprintf("docker run -v %s:/service includeos/build:dev", path)
	if output, err := ExternalCommand(cmdString); err != nil {
		return fmt.Errorf("build failed: %s, error: %v", output, err)
	}
	return nil
}
