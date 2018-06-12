package environment

import (
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
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
}

func (c *SSHClients) getClientByInt(num int) (string, error) {
	switch num {
	case 1:
		return c.Client1, nil
	case 2:
		return c.Client2, nil
	case 3:
		return c.Client3, nil
	default:
		return "", fmt.Errorf("Client num %d does not exist", num)
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

func VerifyEnv(env *Environment) error {
	return nil
}

func runSSHScript(file, SSHRemote string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
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
