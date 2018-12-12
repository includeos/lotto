package mothership

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/mnordsletten/lotto/environment"
	"github.com/sirupsen/logrus"
)

// pushNacl appends lotto-test- to naclfilename and pushes it to mothership
func (m *Mothership) pushNacl(naclFileName string) (string, error) {
	logrus.Infof("Pushing NaCl: %s", naclFileName)
	fileName := path.Base(naclFileName)
	targetName := fmt.Sprintf("lotto-test-%s", strings.TrimSuffix(fileName, filepath.Ext(fileName)))
	request := fmt.Sprintf("push-nacl %s %s --name %s -o id", naclFileName, m.BuilderID, targetName)
	response, err := m.bin(request)
	if err != nil {
		return "", fmt.Errorf("could not push nacl %s: %v", naclFileName, err)
	}
	return response, nil
}

// DeleteNacl takes a naclID and deletes it
func (m *Mothership) DeleteNacl(naclID string) error {
	logrus.Debugf("Deleting NaCl: %s", naclID)
	request := fmt.Sprintf("delete-nacl %s", naclID)
	_, err := m.bin(request)
	if err != nil {
		return fmt.Errorf("could not delete nacl %s: %v", naclID, err)
	}
	return nil
}

// pushUplink deletes any existing uplink with same name and pushes a new one
func (m *Mothership) pushUplink(file, name string) error {
	// Delete any existing uplink if it exists
	request := fmt.Sprintf("inspect-uplink %s", name)
	if _, err := m.bin(request); err == nil {
		request = fmt.Sprintf("delete-uplink %s", name)
		if _, err = m.bin(request); err != nil {
			return fmt.Errorf("couldn't remove uplink %s: %v", name, err)
		}
	}
	// Push the uplink
	request = fmt.Sprintf("push-uplink -o id %s", file)
	uplinkID, err := m.bin(request)
	if err != nil {
		return fmt.Errorf("couldn't push uplink %s: %v", file, err)
	}
	m.uplinkID = uplinkID
	return nil
}

// build will build with the specified naclID that needs to exist on mothership
func (m *Mothership) build(naclID string) (string, error) {
	logrus.Infof("Building image with nacl: %s", naclID)
	m.lastBuildTag = fmt.Sprintf("lotto-%s", time.Now().Format("20060102150405"))
	request := fmt.Sprintf("build --waitAndPrint -n %s -u %s --tag %s Starbase %s", naclID, m.uplinkID, m.lastBuildTag, m.BuilderID)
	checksum, err := m.bin(request)
	if err != nil {
		return "", fmt.Errorf("error building: %v", err)
	}

	return checksum, nil
}

// pullImage saves an image with targetName
func (m *Mothership) pullImage(checksum, targetName string) error {
	logrus.Debugf("Pulling down image: %s", checksum)
	_, err := m.bin(fmt.Sprintf("pull-image %s %s", checksum, targetName))
	if err != nil {
		return fmt.Errorf("error pulling image: %v", err)
	}
	return nil
}

// DeleteImage takes a checksum and deletes it
func (m *Mothership) DeleteImage(imageChecksum string) error {
	logrus.Debugf("Deleting image: %s", imageChecksum)
	request := fmt.Sprintf("delete-image %s", imageChecksum)
	_, err := m.bin(request)
	if err != nil {
		return fmt.Errorf("could not delete image %s: %v", imageChecksum, err)
	}
	return nil
}

// deploy takes an image checksum and deploys to starbase
func (m *Mothership) deploy(checksum string) error {
	logrus.Infof("Deploying %s to %s", checksum, m.Alias)
	request := fmt.Sprintf("deploy --wait %s %s", m.Alias, checksum)
	if _, err := m.bin(request); err != nil {
		return fmt.Errorf("error deploying: %v", err)
	}
	return nil
}

// setAlias takes an ID and gives it the supplied alias
func (m *Mothership) setAlias(alias, ID string) error {
	request := fmt.Sprintf("instance-alias %s %s", ID, alias)
	if _, err := m.bin(request); err != nil {
		return err
	}
	return nil
}

// deleteInstanceByAlias deletes the given instance from mothership
func (m *Mothership) deleteInstanceByAlias(alias string) error {
	request := fmt.Sprintf("delete-instance %s", alias)
	if _, err := m.bin(request); err != nil {
		return err
	}
	return nil
}

// Launch uses requires the environment to specify the launch options
// to launch an instance where it is required.
func (m *Mothership) Launch(imageName string, env environment.Environment) error {
	options := env.LaunchCmdOptions(imageName)
	cmd := exec.Command(m.Binary, options...)
	logrus.Debugf("Launch command: %v", cmd.Args)
	cmd.Env = append(os.Environ())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error running launch cmd: %s, %v", string(output), err)
	}
	return nil
}

// PushImage pushes the supplied image to the mothership
func (m *Mothership) PushImage(imagePath string) (string, error) {
	request := fmt.Sprintf("push-image %s", imagePath)
	var output string
	var err error
	if output, err = m.bin(request); err != nil {
		return "", err
	}
	return output, nil
}

// ServerVersion returns the servers mothership version string
func (m *Mothership) ServerVersion() (string, error) {
	request := fmt.Sprintf("server-version -o json")
	type version struct {
		Version string `json:"Version"`
	}
	output, err := m.bin(request)
	if err != nil {
		return output, err
	}
	var v version
	if err = json.Unmarshal([]byte(output), &v); err != nil {
		return output, err
	}
	if len(v.Version) == 0 {
		logrus.Warningf("Mothership server version is empty")
	}
	return v.Version, nil
}

// StarbaseVersion returns the IncludeOS version that the starbase is using
func (m *Mothership) StarbaseVersion() (string, error) {
	request := fmt.Sprintf("inspect-instance %s -o json", m.Alias)
	output, err := m.bin(request)
	if err != nil {
		return output, err
	}
	var star starbase
	if err := json.Unmarshal([]byte(output), &star); err != nil {
		return output, err
	}
	if len(star.Version) == 0 {
		logrus.Warning("Starbase version is empty")
	}
	return star.Version, nil
}

// BobProvidersUpdate will update all bobProviders
func (m *Mothership) BobProvidersUpdate() error {
	request := fmt.Sprintf("bob update")
	_, err := m.bin(request)
	return err
}

// BobsList list all bobs from all providers and return the output as json
func (m *Mothership) BobsList() (string, error) {
	request := fmt.Sprintf("bob list -o json")
	output, err := m.bin(request)
	if err != nil {
		return output, err
	}
	return output, nil
}

// BobPrepare takes a BobID and a providerID and tries to prepare the Bob
func (m *Mothership) BobPrepare(ID, providerID string) error {
	request := fmt.Sprintf("bob prepare --wait %s %s", providerID, ID)
	if _, err := m.bin(request); err != nil {
		return err
	}
	return nil
}
