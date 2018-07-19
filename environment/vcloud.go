package environment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type Vcloud struct {
	name         string
	Host         string     `json:"host,omitempty"`
	Org          string     `json:"org"`
	Username     string     `json:"username"`
	Password     string     `json:"password"`
	TemplateName string     `json:"templatename"`
	Catalog      string     `json:"catalog"`
	NetworkName  string     `json:"networkname"`
	UplinkFile   string     `json:"uplinkfile"`
	Clients      SSHClients `json:"sshclients"`
	Mothership   string     `json:"mothership"`
}

func (v *Vcloud) SetName(name string) {
	v.name = name
}

func (v *Vcloud) Name() string {
	return v.name
}

func (v *Vcloud) Create() error {
	logrus.Debugf("Creating Vcloud environment. with: %s", v)
	/*
		vappName := fmt.Sprintf("lotto_%s", v.Username)
			createCmd := fmt.Sprintf("/bin/sh -c 'vcd login %s %s %s -p %s; "+
				"vcd vapp create -t %s -c %s %s'",
				v.Host, v.Org, v.Username, v.Password, v.TemplateName, v.Catalog, vappName)
	*/
	cmd := exec.Command("docker", "run", "--rm", "mnordsletten/vcd-cli", "/bin/sh", "-c", "echo Logged in; echo set up environment")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error creating environment: %v", err)
	}

	logrus.Debugf("Successfully set up environment: %s", string(output))
	return nil
}

func (v *Vcloud) Delete() error {
	logrus.Debugf("Deleting Vcloud environment. with: %s", v)
	return nil
}

func (v *Vcloud) GetUplinkFileName() (string, string) {
	name := strings.TrimSuffix(v.UplinkFile, ".json")
	return name, v.UplinkFile
}

func (v *Vcloud) GetUplinkInfo() (UplinkInfo, error) {
	u := UplinkInfo{}
	u.FileName = v.UplinkFile
	u.Name = strings.TrimSuffix(v.UplinkFile, ".json")
	contents, err := ioutil.ReadFile(v.UplinkFile)
	if err != nil {
		return u, fmt.Errorf("error reading file: %v", err)
	}

	type uplink struct {
		Tag string `json:"tag"`
	}

	type uplinkData struct {
		Uplink uplink `json:"uplink"`
	}

	var data uplinkData
	if err := json.Unmarshal(contents, &data); err != nil {
		return u, fmt.Errorf("error decoding JSON: %v", err)
	}

	u.Tag = data.Uplink.Tag
	return u, nil
}

func (v *Vcloud) LaunchCmdOptions(imageName string) []string {
	return []string{"launch", "--hypervisor", "vcloud",
		"--vcloud-vapp", "test-vapp",
		"--vcloud-net", v.NetworkName,
		"--vcloud-address", v.Host,
		"--vcloud-org", v.Org,
		imageName}
}

func (v *Vcloud) RunClientCmd(clientNum int, cmd string) (string, error) {
	clientStr, err := v.Clients.GetClientByInt(clientNum)
	if err != nil {
		return "", fmt.Errorf("error getting client: %v", err)
	}
	return runRemoteCmd(cmd, clientStr)
}

// RunClientCmdScript takes a file and runs it on the remote machine
// It does this by opening an ssh connection and passing in the script contents
// to the running bash process
func (v *Vcloud) RunClientCmdScript(clientNum int, file string) ([]byte, error) {
	logrus.Debugf("Running client script: %s. with: %s", file, v)
	clientStr, err := v.Clients.GetClientByInt(clientNum)
	if err != nil {
		return nil, fmt.Errorf("error getting client: %v", err)
	}
	return runSSHScript(file, clientStr)
}

func (v *Vcloud) GetMothershipName() string {
	return v.Mothership
}
