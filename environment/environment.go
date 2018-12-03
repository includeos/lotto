package environment

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mnordsletten/lotto/util"
	"github.com/sirupsen/logrus"
)

type UplinkInfo struct {
	Name     string
	FileName string
	Tag      string
}

type SSHClients struct {
	Client1 string `json:"client1"`
	Client2 string `json:"client2"`
	Client3 string `json:"client3"`
	Client4 string `json:"client4"`
}

// GetClientByInt returns the contents for a client specified by number
func (c *SSHClients) GetClientByInt(num int) (string, error) {
	switch num {
	case 1:
		return c.Client1, nil
	case 2:
		return c.Client2, nil
	case 3:
		return c.Client3, nil
	case 4:
		return c.Client4, nil
	default:
		return "", fmt.Errorf("Client num %d does not exist", num)
	}
}

type Environment interface {
	SetName(string)
	Name() string
	Create() error
	Delete() error
	GetUplinkInfo() (UplinkInfo, error)
	LaunchCmdOptions(string) []string
	RunClientCmd(clientNum int, cmd string) (string, error)
	RunClientCmdScript(clientNum int, file string) ([]byte, error)
	GetMothershipName() string
}

func VerifyEnv(env Environment) error {
	net1Route := "10.100.0.128/25 via 10.100.0.30 dev $(ip route list scope link | awk '/10.100.0./ {print $3}')"
	net2Route := "10.100.0.0/25 via 10.100.0.140 dev $(ip route list scope link | awk '/10.100.0./ {print $3}')"
	if err := verifyRoute(env, net1Route, 1); err != nil {
		return err
	}
	if err := verifyRoute(env, net1Route, 2); err != nil {
		return err
	}
	if err := verifyRoute(env, net2Route, 3); err != nil {
		return err
	}
	if err := verifyRoute(env, net2Route, 4); err != nil {
		return err
	}
	return nil
}

// verifyRoute will check if route exists, and if not create a new one
func verifyRoute(env Environment, route string, clientNum int) error {
	// Check if route exists already
	existsCmd := fmt.Sprintf("ip route show %s | wc -l", route)
	numLines, err := env.RunClientCmd(clientNum, existsCmd)
	if err != nil {
		return fmt.Errorf("error checking if route to client %d exists: %v", clientNum, err)
	}
	lines, err := strconv.Atoi(numLines)
	if err != nil {
		return fmt.Errorf("could not convert numLines to int: %v", err)
	}
	if lines > 0 {
		logrus.Debugf("Routes exist for client: %d", clientNum)
		return nil
	}

	// set up new route
	newRouteCmd := fmt.Sprintf("sudo ip route add %s", route)
	_, err = env.RunClientCmd(clientNum, newRouteCmd)
	if err != nil {
		return fmt.Errorf("error setting up route: %v", err)
	}
	logrus.Debugf("Set up new routes for client: %d", clientNum)
	return nil
}

func runSSHScript(file, SSHRemote string) ([]byte, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %v", file, err)
	}
	out, err := util.ExternalCommandInput(string(b), []string{"ssh", "-o", "StrictHostKeyChecking=no", SSHRemote})
	if err != nil {
		return nil, fmt.Errorf("problem running external command: %v", err)
	}
	return out, nil
}

func runRemoteCmd(cmd, sshRemote string) (string, error) {
	x := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", sshRemote, cmd)
	byteOutput, err := x.Output()
	if err != nil {
		return string(byteOutput), fmt.Errorf("error running cmd: %v", err)
	}
	output := strings.TrimSuffix(string(byteOutput), "\n")
	return output, nil
}
