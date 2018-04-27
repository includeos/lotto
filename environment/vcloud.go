package environment

import (
	"github.com/mnordsletten/lotto/mothership"
	"github.com/sirupsen/logrus"
)

type Vcloud struct {
	settings EnvSettings
}

func NewVcloud(settings EnvSettings) *Vcloud {
	return &Vcloud{settings}
}

func (v *Vcloud) Name() string {
	return "Vcloud"
}

func (v *Vcloud) Create() error {
	logrus.Debugf("Creating Vcloud environment. with: %s", v.settings)
	return nil
}

func (v *Vcloud) Delete() error {
	logrus.Debugf("Deleting Vcloud environment. with: %s", v.settings)
	return nil
}

func (v *Vcloud) BootStarbase() (error, *mothership.Starbase) {
	logrus.Debugf("Booting starbase in Vcloud. with: %s", v.settings)
	return nil, &mothership.Starbase{ID: "myID", Ip: "myIp"}
}

func (v *Vcloud) RunClientCmd(cmd string) (string, error) {
	logrus.Debugf("Running client cmd: %s. with: %s", cmd, v.settings)
	return "OK", nil
}
