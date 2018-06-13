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
	if err := m.deleteInstanceByAlias(m.starbaseid); err != nil {
		logrus.Debugf("%s could not be removed: %v", m.starbaseid, err)
	}

	// Create a new tag which is inserted in the image that is built
	m.tag = fmt.Sprintf("lotto-%s", time.Now().Format("20060102150405"))
	if err := m.createCleanStarbase(); err != nil {
		return err
	}

	// Launch the newly built starbase
	if err := m.Launch(cleanStarbaseImage, env); err != nil {
		return err
	}

	// Wait until the starbase with the "unique" tag we just created connects
	if err := m.waitUntilStarbaseConnects(m.tag); err != nil {
		return err
	}

	// Set the alias to something generic
	/*
		if err := m.setAlias(m.starbaseid, m.tag); err != nil {
			return err
		}
	*/
	return nil
}

func (m *Mothership) CheckStarbaseIDInUse() bool {
	type starbase struct {
		Online bool
	}
	request := fmt.Sprintf("inspect-instance %s -o json", m.starbaseid)
	response, err := m.bin(request)
	if err != nil {
		logrus.Infof("Mothership: starbase %s does not exist", m.starbaseid)
		return false
	}
	logrus.Infof("Mothership: starbase %s exists", m.starbaseid)
	var star starbase
	if err := json.Unmarshal([]byte(response), &star); err != nil {
		logrus.Debugf("could not check if starbase is online: %v", err)
		return false
	}
	if !star.Online {
		logrus.Infof("Mothership: starbase %s is offline", m.starbaseid)
		return false
	}
	return true
}

func (m *Mothership) waitUntilStarbaseConnects(tag string) error {
	logrus.Debugf("Now waiting for starbase %s to connect", tag)
	deadLine := time.Now().Add(180 * time.Second)
	for time.Now().Before(deadLine) {
		req := fmt.Sprintf("inspect-instance %s", m.starbaseid)
		if _, err := m.bin(req); err == nil {
			logrus.Debugf("Starbase %s connected", m.starbaseid)
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("Starbase %s never connected to mothership", tag)
}
