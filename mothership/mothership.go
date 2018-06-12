package mothership

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/mnordsletten/lotto/environment"
	"github.com/sirupsen/logrus"
)

const cleanStarbaseImage = "clean-starbase"

type Mothership struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	NoTLS      bool   `json:"notls,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	VerifyTLS  bool   `json:"verifytls,omitempty"`
	Binary     string `json:"binarypath,omitempty"`
	uplinkname string
	starbaseid string
	tag        string
}

// NewMothership is used to generate a Mothership struct.
func NewMothership(host, username, password, binary string, port int, notls, verifytls bool, env environment.Environment) (*Mothership, error) {
	m := &Mothership{Host: host, Port: port, NoTLS: notls, VerifyTLS: verifytls, Username: username, Password: password, Binary: binary}

	// Push the uplink to the mothership
	uplinkInfo, err := env.GetUplinkInfo()
	if err != nil {
		return m, fmt.Errorf("error getting uplink info: %v", err)
	}
	if err = m.pushUplink(uplinkInfo.FileName, uplinkInfo.Name); err != nil {
		return m, fmt.Errorf("error pushing uplink %s: %v", uplinkInfo.FileName, err)
	}
	m.uplinkname = uplinkInfo.Name
	m.starbaseid = fmt.Sprintf("lotto-%s", m.Username)
	logrus.Infof("StarbaseID to use: %s", m.starbaseid)
	m.tag = uplinkInfo.Tag
	return m, nil
}

// DeployNacl takes a nacl []byte and deploys it to the Mothership
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

// LaunchCleanStarbase launches a starbase in the specified environment and connects it
// to the Mothership.
func (m *Mothership) LaunchCleanStarbase(env environment.Environment) error {
	// Remove old alias, if it does not exist an error is returned but that does not matter
	if err := m.removeAlias(m.starbaseid); err != nil {
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

func (m *Mothership) pushNacl(naclFileName string) (string, error) {
	logrus.Infof("Pushing NaCl: %s", naclFileName)
	targetName := strings.Split(naclFileName, ".")[0]
	request := fmt.Sprintf("push-nacl %s --name %s -o id", naclFileName, targetName)
	response, err := m.bin(request)
	if err != nil {
		return "", fmt.Errorf("could not push nacl %s: %v", naclFileName, err)
	}
	return response, nil
}

func (m *Mothership) build(naclID string) (string, error) {
	logrus.Infof("Building image with nacl: %s", naclID)
	request := fmt.Sprintf("build --waitAndPrint -n %s -u %s --tag %s Starbase", naclID, m.uplinkname, m.tag)
	checksum, err := m.bin(request)
	if err != nil {
		return "", fmt.Errorf("error building: %v", err)
	}

	return checksum, nil
}

func (m *Mothership) pullImage(checksum, targetName string) error {
	logrus.Debugf("Pulling down image: %s", checksum)
	_, err := m.bin(fmt.Sprintf("pull-image %s %s", checksum, targetName))
	if err != nil {
		return fmt.Errorf("error pulling image: %v", err)
	}
	return nil
}

func (m *Mothership) deploy(checksum string) error {
	logrus.Infof("Deploying %s to %s", checksum, m.starbaseid)
	request := fmt.Sprintf("deploy %s %s", m.starbaseid, checksum)
	if _, err := m.bin(request); err != nil {
		return fmt.Errorf("error deploying: %v", err)
	}
	return nil
}

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

func (m *Mothership) setAlias(alias, ID string) error {
	request := fmt.Sprintf("instance-alias %s %s", ID, alias)
	if _, err := m.bin(request); err != nil {
		return err
	}
	return nil
}

func (m *Mothership) removeAlias(alias string) error {
	request := fmt.Sprintf("delete-instance %s", alias)
	if _, err := m.bin(request); err != nil {
		return err
	}
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
	request = fmt.Sprintf("push-uplink %s", file)
	if _, err := m.bin(request); err != nil {
		return fmt.Errorf("couldn't push uplink %s: %v", file, err)
	}
	return nil
}

func (m *Mothership) bin(request string) (string, error) {
	tlsFlag := ""
	if m.NoTLS {
		tlsFlag = "--notls"
	}
	var tlsInsecureFlag string
	if !m.VerifyTLS {
		tlsInsecureFlag = "--tlsInsecureSkipVerify"
	}
	reqList := strings.Split(request, " ")
	reqList = append(reqList, "--username", m.Username, "--password", m.Password,
		"--host", m.Host, "--port", strconv.Itoa(m.Port), tlsFlag, tlsInsecureFlag)
	reqList = removeEmptyInSlice(reqList)
	cmd := exec.Command(m.Binary, reqList...)
	logrus.Debugf("mothership command: %v", cmd.Args)

	byteOutput, err := cmd.CombinedOutput()
	output := strings.TrimSuffix(string(byteOutput), "\n")
	if err != nil {
		return "", fmt.Errorf("error running mothership cmd: %s: %s, %v", request, string(output), err)
	}

	return string(output), nil
}

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

func removeEmptyInSlice(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
