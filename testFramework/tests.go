package testFramework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"path"
	"time"

	"github.com/mnordsletten/lotto/environment"
	"github.com/mnordsletten/lotto/mothership"
	"github.com/mnordsletten/lotto/util"
	"github.com/sirupsen/logrus"
)

type TestConfig struct {
	ID                  string                 `json:"id"`
	NaclFile            string                 `json:"naclfile"`
	ClientCommandScript string                 `json:"clientcommandscript"`
	HostCommandScript   string                 `json:"hostcommandscript"`
	Setup               environment.SSHClients `json:"setup"`
	Level1              int                    `json:"level1"`
	Level2              int                    `json:"level2"`
	Level3              int                    `json:"level3"`
	ShouldFail          bool                   `json:"shouldfail"`
	CustomServicePath   string                 `json:"customservicepath"`
	Deploy              bool                   `json:"deploy"`
	ImageID             string
	testPath            string
}

type TestResult struct {
	Time              string  // Time that test results were recorded
	Name              string  // Name of test
	Sent              int     // Total number of requests sent
	Received          int     // Total number of replies received
	Rate              float32 // Requests pr second
	Avg               float32 // Average response time
	SuccessPercentage float32 // Percentage of packets that pass
	Raw               string  // Raw output from the command
	ShouldFail        bool    // If the test expects to fail
}

type HostCommandTemplate struct {
	MothershipBinPathAndName string
	OriginalAlias            string
	ImageID                  string
}

func (tr TestResult) String() string {
	return fmt.Sprintf("Percentage: %.1f%%, sent/recv: %d/%d, ShouldFail: %t, Name: %s", tr.SuccessPercentage, tr.Sent, tr.Received, tr.ShouldFail, tr.Name)
}

// RunTest runs the clientCmdScript on either host or client1 level number of times and returns a TestResult
func (t *TestConfig) RunTest(level int, env environment.Environment, mother *mothership.Mothership) (TestResult, error) {
	if err := t.prepareTest(env); err != nil {
		return TestResult{}, fmt.Errorf("error preparing test: %v", err)
	}
	logrus.Infof("Starting test: %s", path.Base(t.testPath))
	var results []TestResult
	for i := 0; i < level; i++ {
		var testOutput []byte
		var testResult TestResult
		var err error
		// Run test either in client or in host
		if t.ClientCommandScript != "" {
			if testOutput, err = t.runClientTest(env); err != nil {
				return testResult, fmt.Errorf("could not run client test: %v", err)
			}
		} else if t.HostCommandScript != "" {
			if testOutput, err = t.runHostTest(mother); err != nil {
				return testResult, fmt.Errorf("could not run lotto test: %v", err)
			}
		} else {
			return testResult, fmt.Errorf("no testFile found")
		}

		// Parse test results
		if err = json.Unmarshal(testOutput, &testResult); err != nil {
			return testResult, fmt.Errorf("could not parse testResults: %v", err)
		}
		testResult.Time = time.Now().Format(time.RFC3339)
		testResult.Name = path.Base(t.testPath)

		// Calculate success
		testResult.ShouldFail = t.ShouldFail
		if t.ShouldFail {
			if testResult.Received > 0 {
				testResult.SuccessPercentage = 0
			} else {
				testResult.SuccessPercentage = 100
			}
		} else {
			testResult.SuccessPercentage = float32(testResult.Received) / float32(testResult.Sent) * 100
		}
		results = append(results, testResult)
	}
	return combineTestResults(results), nil
}

func (t *TestConfig) runClientTest(env environment.Environment) ([]byte, error) {
	testOutput, err := env.RunClientCmdScript(1, t.ClientCommandScript)
	if err != nil {
		return testOutput, fmt.Errorf("could not run client command script: %v", err)
	}
	return testOutput, nil
}

func (t *TestConfig) runHostTest(mother *mothership.Mothership) ([]byte, error) {
	// Process script file as template, replace template objects with actual info.
	f, err := ioutil.ReadFile(t.HostCommandScript)
	if err != nil {
		return nil, fmt.Errorf("error reading lotto test template: %v", err)
	}

	// Parse requires a string
	m, err := template.New("test").Parse(string(f))
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %v", err)
	}

	templ := HostCommandTemplate{MothershipBinPathAndName: mother.CLICommand(),
		OriginalAlias: mother.Alias,
		ImageID:       t.ImageID}
	var script bytes.Buffer
	if err = m.Execute(&script, templ); err != nil {
		return nil, fmt.Errorf("error executing template: %v", err)
	}

	out, err := util.ExternalCommandInput(html.UnescapeString(script.String()), nil)
	if err != nil {
		fmt.Printf("problem with external: %v", err)
	}
	// Unmarshal test results into testResult
	return out, nil
}

func (t *TestConfig) prepareTest(env environment.Environment) error {
	t.Setup.RunFuncOnAllClients(func(input string) string {
		return path.Join(t.testPath, path.Base(input))
	})
	t.Setup.PopulateSlice()
	for i, script := range t.Setup.ClientSlice {
		if script != "" {
			if output, err := env.RunClientCmdScript(i+1, script); err != nil {
				return fmt.Errorf("Could not run ClientCmdScript: %s: %v", string(output), err)
			}
		}
	}
	return nil
}

func combineTestResults(results []TestResult) TestResult {
	end := TestResult{}
	for _, result := range results {
		end.Name = result.Name
		end.Sent += result.Sent
		end.Received += result.Received
		end.Rate += result.Rate
		end.ShouldFail = result.ShouldFail
	}
	if end.ShouldFail {
		if end.Received > 0 {
			end.SuccessPercentage = 0
		} else {
			end.SuccessPercentage = 100
		}
	} else {
		end.SuccessPercentage = float32(end.Received) / float32(end.Sent) * 100
	}
	end.Time = time.Now().Format(time.RFC3339)
	return end
}
