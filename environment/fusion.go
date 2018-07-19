package environment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/sirupsen/logrus"
)

type Fusion struct {
	name       string
	Clients    SSHClients `json:"sshclients"`
	UplinkFile string     `json:"uplinkFile"`
	VmspecPath string     `json:"vmSpecPath"`
	Mothership string     `json:"mothership"`
}

func NewFusion(clients SSHClients, uplinkFilePath, vmSpecPath, mothership string) *Fusion {
	f := &Fusion{}
	f.Clients = clients
	f.UplinkFile = uplinkFilePath
	f.VmspecPath = vmSpecPath
	f.Mothership = mothership
	// TODO: Add checks for all of these, that they are supplied and that they work
	return f
}

func (f *Fusion) SetName(name string) {
	f.name = name
}

func (f *Fusion) Name() string {
	return f.name
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

func (f *Fusion) RunClientCmd(clientNum int, cmd string) (string, error) {
	clientStr, err := f.Clients.GetClientByInt(clientNum)
	if err != nil {
		return "", fmt.Errorf("error getting client: %v", err)
	}
	return runRemoteCmd(cmd, clientStr)
}

func (f *Fusion) RunClientCmdScript(clientNum int, file string) ([]byte, error) {
	clientStr, err := f.Clients.GetClientByInt(clientNum)
	if err != nil {
		return nil, fmt.Errorf("error getting client: %v", err)
	}
	logrus.Debugf("Running client script: %s. with: %s", file, clientStr)
	return runSSHScript(file, clientStr)
}

func (f *Fusion) GetMothershipName() string {
	return f.Mothership
}
