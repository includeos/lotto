package util

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"
)

// MyError is an Error implementation
type MyError struct {
	Cmd        string
	Stdout     string
	Stderr     string
	CommandErr error
}

func (e MyError) Error() string {
	return fmt.Sprintf(
		"Command run: %s\n"+
			"Stdout: %s\n"+
			"Stderr: %s\n"+
			"Error: %s\n",
		e.Cmd, e.Stdout, e.Stderr, e.CommandErr)
}

func splitString(cmdString string) *exec.Cmd {
	// Process input string
	cmdString = strings.Join(strings.Fields(cmdString), " ")
	cmdList := []string(strings.Split(cmdString, " "))
	return exec.Command(cmdList[0], cmdList[1:]...)
}

// ExternalCommand runs cmdString and returns output
func ExternalCommand(cmdString string) (string, error) {
	var errbuf bytes.Buffer
	cmd := splitString(cmdString)
	cmd.Stderr = &errbuf

	// Run output command
	output, err := cmd.Output()
	stdout := string(output[:])
	if err != nil {
		return stdout, MyError{cmdString, stdout, errbuf.String(), err}
	}
	return stdout, nil
}

// ExternalCommandBackground starts command as a background process
func ExternalCommandBackground(cmdString string) (*exec.Cmd, error) {
	var errbuf bytes.Buffer
	cmd := splitString(cmdString)
	cmd.Stderr = &errbuf

	if err := cmd.Start(); err != nil {
		return cmd, MyError{cmdString, "", errbuf.String(), err}
	}
	return cmd, nil
}

// ExternalCommandTimeout runs cmdString with timeout and returns output
func ExternalCommandTimeout(cmdString string, timeout time.Duration) (string, error) {
	var errbuf, outbuf bytes.Buffer
	cmd := splitString(cmdString)
	cmd.Stderr = &errbuf
	cmd.Stdout = &outbuf

	// Run command
	if err := cmd.Start(); err != nil {
		return outbuf.String(), MyError{cmdString, outbuf.String(), errbuf.String(), err}
	}

	// Set up timeout function
	if timeout > 0 {
		timer := time.AfterFunc(timeout, func() {
			cmd.Process.Kill()
		})
		if err := cmd.Wait(); err != nil {
			return outbuf.String(), MyError{cmdString, outbuf.String(), errbuf.String(), err}
		}
		timer.Stop()
	} else {
		if err := cmd.Wait(); err != nil {
			return outbuf.String(), MyError{cmdString, outbuf.String(), errbuf.String(), err}
		}
	}

	// Return output
	return strings.TrimSpace(outbuf.String()), nil
}

// ExternalCommandCombinedOutput runs cmdString and returns both stdout and stderr
func ExternalCommandCombinedOutput(cmdString string) (string, error) {
	cmd := splitString(cmdString)

	// Run output command
	combinedOutput, err := cmd.CombinedOutput()
	if err != nil {
		return string(combinedOutput), MyError{cmdString, "", string(combinedOutput), err}
	}
	return string(combinedOutput), nil
}

// ExternalCommandInput takes an input that is sent to bash -s stdin
func ExternalCommandInput(input string, sshRemote []string) ([]byte, error) {
	var command []string
	if len(sshRemote) > 0 {
		command = append(sshRemote, "bash -s")
	} else {
		command = []string{"bash", "-s"}
	}
	x := exec.Command(command[0], command[1:]...)
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
	_, err = io.WriteString(stdin, input)
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
