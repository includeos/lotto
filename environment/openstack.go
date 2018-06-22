package environment

import (
	"github.com/sirupsen/logrus"
)

type Openstack struct {
	Mothership string
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

func (o *Openstack) RunClientCmd(clientNum int, cmd string) (string, error) {
	logrus.Debugf("Running client cmd: %s", cmd)
	return "OK", nil
}

func (o *Openstack) RunClientCmdScript(clientNum int, file string) ([]byte, error) {
	return nil, nil
}

func (o *Openstack) GetMothershipName() string {
	return o.Mothership
}
