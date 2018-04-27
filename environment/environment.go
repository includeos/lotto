package environment

import (
	"fmt"

	"github.com/mnordsletten/lotto/mothership"
)

type EnvSettings struct {
	Username string `json:"username"`
	Password string `json:"password"`
	AppName  string `json:"appname"`
	Address  string `json:"address"`
}

func (s EnvSettings) String() string {
	return fmt.Sprintf("%s:%s@%s/vapp=%s", s.Username, s.Password, s.Address, s.AppName)
}

type Environment interface {
	Name() string
	Create() error
	Delete() error
	BootStarbase() (error, *mothership.Starbase)
	RunClientCmd(cmd string) (string, error)
}
