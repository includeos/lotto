package mothership

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mnordsletten/lotto/environment"
	"github.com/sirupsen/logrus"
)

// DeployNacl takes a nacl file name and deploys it to the Mothership
func (m *Mothership) DeployNacl(naclFileName string) error {
	naclID, err := m.pushNacl(naclFileName)
	if err != nil {
		return err
	}

	checksum, err := m.build(naclID)
	if err != nil {
		return err
	}
	if err := m.deploy(checksum); err != nil {
		return err
	}
	return nil
}

// createCleanStarbase uses a standard nacl to build and pull down the image
func (m *Mothership) createCleanStarbase() error {
	naclFileName := "clean-starbase.nacl"
	var naclID, checksum string
	var err error
	// Push-NaCl
	if naclID, err = m.pushNacl(naclFileName); err != nil {
		return fmt.Errorf("error pushing NaCl: %v", err)
	}

	// Build
	if checksum, err = m.build(naclID); err != nil {
		return fmt.Errorf("error building %s: %v", naclFileName, err)
	}

	// Pull image
	if err := m.pullImage(checksum, cleanStarbaseImage); err != nil {
		return fmt.Errorf("error pulling built image: %v", err)
	}

	return nil
}

// LaunchCleanStarbase launches a starbase in the specified environment and connects it
// to the Mothership.
func (m *Mothership) LaunchCleanStarbase(env environment.Environment) error {
	// Remove old alias, if it does not exist an error is returned but that does not matter
	if err := m.deleteInstanceByAlias(m.alias); err != nil {
		logrus.Debugf("%s could not be removed: %v", m.alias, err)
	}

	// Create a new tag which is inserted in the image that is built
	if err := m.createCleanStarbase(); err != nil {
		return err
	}

	// Launch the newly built starbase
	if err := m.Launch(cleanStarbaseImage, env); err != nil {
		return err
	}

	// Wait until the starbase with the "unique" tag we just created connects
	id, err := m.waitUntilStarbaseConnects(m.lastBuildTag)
	if err != nil {
		return err
	}

	// Set the alias to something generic
	if err := m.setAlias(m.alias, id); err != nil {
		return err
	}
	return nil
}

func (m *Mothership) CheckStarbaseIDInUse() bool {
	type starbase struct {
		Status string `json:"status"`
	}
	request := fmt.Sprintf("inspect-instance %s -o json", m.alias)
	response, err := m.bin(request)
	if err != nil {
		logrus.Infof("Mothership: starbase with alias: %s does not exist", m.alias)
		return false
	}
	logrus.Infof("Mothership: starbase with alias: %s exists", m.alias)
	var star starbase
	if err := json.Unmarshal([]byte(response), &star); err != nil {
		logrus.Debugf("could not check if starbase is connected: %v", err)
		return false
	}
	if star.Status != "connected" {
		logrus.Infof("Mothership: starbase with alias: %s is disconnected", m.alias)
		return false
	}
	return true
}

func (m *Mothership) waitUntilStarbaseConnects(tag string) (string, error) {
	logrus.Debugf("Now waiting for starbase with tag: %s to connect", tag)
	ids := make(map[string]interface{})
	deadLine := time.Now().Add(180 * time.Second)
	for time.Now().Before(deadLine) {
		req := fmt.Sprintf("search %s --instancefilter tag -o json", tag)
		if output, err := m.bin(req); err == nil {
			// Cmd did not return an error
			err := json.Unmarshal([]byte(output), &ids)
			if err != nil {
				return "", fmt.Errorf("error unmarshaling json from search: %v", err)
			}
			for key := range ids {
				// there is at least 1 key in the output
				logrus.Debugf("Starbase with tag: %s connected", tag)
				return key, nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return "", fmt.Errorf("Starbase %s never connected to mothership", tag)
}
