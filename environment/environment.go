package environment

import (
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type UplinkInfo struct {
	Name     string
	FileName string
	Tag      string
}

type SSHClients struct {
	Client1     string `json:"client1"`
	Client2     string `json:"client2"`
	Client3     string `json:"client3"`
	Client4     string `json:"client4"`
	ClientSlice []string
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

func (c *SSHClients) PopulateSlice() {
	c.ClientSlice = make([]string, 4)
	c.ClientSlice[0] = c.Client1
	c.ClientSlice[1] = c.Client2
	c.ClientSlice[2] = c.Client3
	c.ClientSlice[3] = c.Client4
}

// RunFuncOnAllClients takes a function and updates the strings for all existing clients
func (c *SSHClients) RunFuncOnAllClients(f func(string) string) {
	if c.Client1 != "" {
		c.Client1 = f(c.Client1)
	}
	if c.Client2 != "" {
		c.Client2 = f(c.Client2)
	}
	if c.Client3 != "" {
		c.Client3 = f(c.Client3)
	}
	if c.Client4 != "" {
		c.Client4 = f(c.Client4)
	}
}

type EnvSettings struct {
	Vcloud Vcloud
	Fusion Fusion
}

type Environment interface {
	Name() string
	Create() error
	Delete() error
	GetUplinkInfo() (UplinkInfo, error)
	LaunchCmdOptions(string) []string
	RunClientCmd(clientNum int, cmd string) (string, error)
	RunClientCmdScript(clientNum int, file string) ([]byte, error)
}

func VerifyEnv(env Environment) error {
	net1Route := "10.100.0.128/25 via 10.100.0.30 dev ens38"
	net2Route := "10.100.0.0/25 via 10.100.0.140 dev ens38"
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
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %v", file, err)
	}
	x := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", SSHRemote, "bash -s")
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

func runRemoteCmd(cmd, sshRemote string) (string, error) {
	x := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", sshRemote, cmd)
	byteOutput, err := x.Output()
	if err != nil {
		return string(byteOutput), fmt.Errorf("error running cmd: %v", err)
	}
	output := strings.TrimSuffix(string(byteOutput), "\n")
	return output, nil
}
