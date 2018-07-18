package mothership

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/mnordsletten/lotto/environment"
	"github.com/sirupsen/logrus"
)

const cleanStarbaseImage = "clean-starbase"

// Mothership defines all options necessary to keep track of the mothership and
// the starbase that is connected to it
type Mothership struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	NoTLS          bool   `json:"notls,omitempty"`
	Username       string `json:"username,omitempty"`
	Password       string `json:"password,omitempty"`
	VerifyTLS      bool   `json:"verifytls,omitempty"`
	Binary         string `json:"binarypath,omitempty"`
	uplinkname     string
	uplinkFileName string
	Alias          string
	lastBuildTag   string
	lastCheckTime  time.Time
}

// NewMothership is used to generate a Mothership struct.
func NewMothership(host, username, password, binary string, port int, notls, verifytls bool, env environment.Environment) (*Mothership, error) {
	m := &Mothership{Host: host, Port: port, NoTLS: notls, VerifyTLS: verifytls, Username: username, Password: password, Binary: binary}

	// Push the uplink to the mothership
	uplinkInfo, err := env.GetUplinkInfo()
	if err != nil {
		return m, fmt.Errorf("error getting uplink info: %v", err)
	}
	if err = m.pushUplink(uplinkInfo.FileName, uplinkInfo.Name); err != nil {
		return m, fmt.Errorf("error pushing uplink %s: %v", uplinkInfo.FileName, err)
	}
	m.uplinkname = uplinkInfo.Name
	m.uplinkFileName = uplinkInfo.FileName
	m.Alias = fmt.Sprintf("lotto-%s", m.Username)
	logrus.Infof("Starbase alias to use: %s", m.Alias)

	// lastCheckTime is used to know when the testing started
	m.lastCheckTime = time.Now()
	return m, nil
}

// bin uses the mothership binary CLI to perform all actions towards mothership
func (m *Mothership) bin(request string) (string, error) {
	commandList := strings.Split(m.CLICommand(), " ")
	commandList = append(commandList, strings.Split(request, " ")...)
	cmd := exec.Command(commandList[0], commandList[1:]...)
	logrus.Debugf("mothership command: %v", cmd.Args)

	byteOutput, err := cmd.CombinedOutput()
	output := strings.TrimSuffix(string(byteOutput), "\n")
	if err != nil {
		return "", fmt.Errorf("error running mothership cmd: %s: %s, %v", request, string(output), err)
	}
	return string(output), nil
}

// CLICommand populates the binary with the correct flags
func (m *Mothership) CLICommand() string {
	tlsFlag := ""
	if m.NoTLS {
		tlsFlag = "--notls"
	}
	var tlsInsecureFlag string
	if !m.VerifyTLS {
		tlsInsecureFlag = "--tlsInsecureSkipVerify"
	}
	bin := fmt.Sprintf("%s --username %s --password %s --host %s --port %d %s %s",
		m.Binary,
		m.Username,
		m.Password,
		m.Host,
		m.Port,
		tlsFlag,
		tlsInsecureFlag)
	return strings.Join(strings.Fields(bin), " ")
}
