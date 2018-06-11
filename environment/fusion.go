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

type Fusion struct {
	SshRemote  string `json:"sshRemote"`
	UplinkFile string `json:"uplinkFile"`
	VmspecPath string `json:"vmSpecPath"`
}

func NewFusion(sshRemote, uplinkFilePath, vmSpecPath string) *Fusion {
	f := &Fusion{}
	f.SshRemote = sshRemote
	f.UplinkFile = uplinkFilePath
	f.VmspecPath = vmSpecPath
	// TODO: Add checks for all of these, that they are supplied and that they work
	return f
}

func (f *Fusion) Name() string {
	return "Fusion"
}

func (f *Fusion) Create() error {
	logrus.Debugf("Creating Fusion environment")
	return nil
}

func (f *Fusion) Delete() error {
	logrus.Debugf("Deleting Fusion environment")
	return nil
}

func (f *Fusion) GetUplinkInfo() (UplinkInfo, error) {
	u := UplinkInfo{}
	u.FileName = f.UplinkFile
	u.Name = strings.TrimSuffix(f.UplinkFile, ".json")
	contents, err := ioutil.ReadFile(f.UplinkFile)
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

func (f *Fusion) LaunchCmdOptions(imageName string) []string {
	return []string{"launch", "--hypervisor", "vmware", "--connectionType", "custom", "--vmjsonpath", f.VmspecPath, imageName}
}

func (f *Fusion) RunClientCmd(cmd string) (string, error) {
	logrus.Debugf("Running client cmd: %s", cmd)

	return "OK", nil
}

func (f *Fusion) RunClientCmdScript(file string) ([]byte, error) {
	logrus.Debugf("Running client script: %s. with: %s", file, f)
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	x := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", f.SshRemote, "bash -s")
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
