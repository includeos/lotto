package environment

import (
	"github.com/sirupsen/logrus"
)

type Openstack struct {
}

func NewOpenstack() *Openstack {
	return &Openstack{}
}

func (o *Openstack) Name() string {
	return "Openstack"
}

func (o *Openstack) Create() error {
	logrus.Debugf("Creating Openstack environment")
	return nil
}

func (o *Openstack) Delete() error {
	logrus.Debugf("Deleting Openstack environment")
	return nil
}

func (o *Openstack) GetUplinkFileName() (string, string) {
	return "", ""
}

func (o *Openstack) GetUplinkInfo() (UplinkInfo, error) {
	return UplinkInfo{}, nil
}

func (o *Openstack) LaunchCmdOptions(string) []string {
	return []string{}
}

func (o *Openstack) RunClientCmd(cmd string) (string, error) {
	logrus.Debugf("Running client cmd: %s", cmd)
	return "OK", nil
}

func (o *Openstack) RunClientCmdScript(file string) ([]byte, error) {
	return nil, nil
}
