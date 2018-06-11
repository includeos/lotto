package environment

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type Vcloud struct {
	Host         string `json:"host,omitempty"`
	Org          string
	Username     string
	Password     string
	TemplateName string
	Catalog      string
	NetworkName  string
	SshRemote    string
	UplinkFile   string
}

func (v *Vcloud) Name() string {
	return "Vcloud"
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

func (v *Vcloud) RunClientCmd(cmd string) (string, error) {
	logrus.Debugf("Running client cmd: %s. with: %s", cmd, v)
	x := exec.Command("ssh", v.SshRemote, cmd)
	byteOutput, err := x.Output()
	if err != nil {
		return "", fmt.Errorf("error running cmd: %v", err)
	}
	output := strings.TrimSuffix(string(byteOutput), "\n")
	return output, nil
}

// RunClientCmdScript takes a file and runs it on the remote machine
// It does this by opening an ssh connection and passing in the script contents
// to the running bash process
func (v *Vcloud) RunClientCmdScript(file string) ([]byte, error) {
	logrus.Debugf("Running client script: %s. with: %s", file, v)
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	x := exec.Command("ssh", v.SshRemote, "bash -s")
	stdin, err := x.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := x.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := x.StderrPipe()
	if err != nil {
		return nil, err
	}
	err = x.Start()
	if err != nil {
		return nil, err
	}

	// Send script contents to process
	_, err = io.WriteString(stdin, string(bytes))
	if err != nil {
		return nil, err
	}
	stdin.Close()
	out, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("stdout read err: %v", err)
	}
	outerr, err := ioutil.ReadAll(stderr)
	if err != nil {
		return nil, fmt.Errorf("stderr read err: %v", err)
	}
	// Check for exit errors
	if err := x.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%v: %s", exiterr, string(outerr))
		}
	}
	return out, nil

}
