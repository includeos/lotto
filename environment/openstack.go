package environment

import (
	"github.com/mnordsletten/lotto/mothership"
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

func (o *Openstack) BootStarbase() (error, *mothership.Starbase) {
	logrus.Debugf("Booting starbase in Openstack")
	return nil, &mothership.Starbase{ID: "myID", Ip: "myIp"}
}

func (o *Openstack) RunClientCmd(cmd string) (string, error) {
	logrus.Debugf("Running client cmd: %s", cmd)
	return "OK", nil
}
