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
	Cleanup             environment.SSHClients `json:"cleanup"`
	ShouldFail          bool                   `json:"shouldfail"`
	CustomServicePath   string                 `json:"customservicepath"`
	NoDeploy            bool                   `json:"nodeploy"`
	SkipRebuild         bool
	ImageID             string
	testPath            string
	Name                string
	NaclFileShasum      string
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
	BuilderID                string
}

func (tr TestResult) String() string {
	return fmt.Sprintf("Result: %.1f%% %d/%d, Name: %s, ShouldFail: %t", tr.SuccessPercentage, tr.Received, tr.Sent, tr.Name, tr.ShouldFail)
}

// RunTest runs the clientCmdScript on either host or client1 level number of times and returns a TestResult
func (t *TestConfig) RunTest(level int, env environment.Environment, mother *mothership.Mothership) (TestResult, error) {
	// Prepare test before run
	if err := t.runScriptsOnClients(env, t.Setup); err != nil {
		return TestResult{}, fmt.Errorf("error preparing test: %v", err)
	}
	defer t.cleanupTest(mother, env)
	var results []TestResult
	for i := 0; i < level; i++ {
		var testOutput []byte
		var testResult TestResult
		testResult.Name = t.Name
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
	logrus.Debugf("Running host test: %s", t.HostCommandScript)
	f, err := ioutil.ReadFile(t.HostCommandScript)
	if err != nil {
		return nil, fmt.Errorf("error reading lotto test template: %v", err)
	}

	// Parse requires a string
	m, err := template.New("test").Parse(string(f))
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %v", err)
	}

	templ := HostCommandTemplate{
		MothershipBinPathAndName: mother.CLICommand(),
		OriginalAlias:            mother.Alias,
		ImageID:                  t.ImageID,
		BuilderID:                mother.BuilderID,
	}
	var script bytes.Buffer
	if err = m.Execute(&script, templ); err != nil {
		return nil, fmt.Errorf("error executing template: %v", err)
	}

	out, err := util.ExternalCommandInput(html.UnescapeString(script.String()), nil)
	if err != nil {
		return out, fmt.Errorf("Host test external command failed: %v", err)
	}
	// Unmarshal test results into testResult
	return out, nil
}

func (t *TestConfig) runScriptsOnClients(env environment.Environment, scripts environment.SSHClients) error {
	for i := 1; i <= 4; i++ {
		scriptName, err := scripts.GetClientByInt(i)
		if err != nil {
			return fmt.Errorf("error getting client%d: %v", i, err)
		} else if scriptName == "" {
			continue
		}
		logrus.Debugf("running script: %s on client%d", scriptName, i)
		scriptPath := path.Join(t.testPath, scriptName)
		if output, err := env.RunClientCmdScript(i, scriptPath); err != nil {
			return fmt.Errorf("error running script %s on client%d: out: %s: %v", scriptPath, i, output, err)
		}
	}
	return nil
}

func (t *TestConfig) cleanupTest(mother *mothership.Mothership, env environment.Environment) {
	// Remove NaCl
	if len(t.NaclFileShasum) > 0 {
		if err := mother.DeleteNacl(t.NaclFileShasum); err != nil {
			logrus.Errorf("could not clean up nacl: %v", err)
		}
	}

	// Remove image
	if len(t.ImageID) > 0 {
		if err := mother.DeleteImage(t.ImageID); err != nil {
			logrus.Errorf("could not clean up image: %v", err)
		}
	}

	if err := t.runScriptsOnClients(env, t.Cleanup); err != nil {
		logrus.Errorf("Error cleaning up: %v", err)
	}
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
