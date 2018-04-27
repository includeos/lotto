package mothership

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type Mothership struct {
	Uplink   string `json:"uplink"`
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (m Mothership) String() string {
	return fmt.Sprintf("./mothership --username %s --password %s --host %s", m.Username, m.Password, m.Address)
}

func NewMothership(uplink, address, username, password string) *Mothership {
	return &Mothership{uplink, address, username, password}
}

func (m *Mothership) PushNacl(nacl string) error {
	logrus.Debugf("Pushing NaCl: %s. with: %s", nacl, m)
	return nil
}

func (m *Mothership) Build(nacl string) error {
	logrus.Debugf("Building with nacl: %s. with: %s", nacl, m)
	return nil
}

func (m *Mothership) Deploy(nacl string, sb *Starbase) error {
	logrus.Debugf("Deploying %s to %s. with: %s", nacl, sb.ID, m)
	return nil
}

func (m *Mothership) DeployNacl(nacl string, sb *Starbase) error {
	if err := m.PushNacl(nacl); err != nil {
		return err
	}
	if err := m.Build(nacl); err != nil {
		return err
	}
	if err := m.Deploy(nacl, sb); err != nil {
		return err
	}
	return nil
}
