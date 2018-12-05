package testFramework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"path"
	"strconv"
	"time"

	"github.com/logrusorgru/aurora"
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
	CustomServicePath   string                 `json:"customservicepath"`
	NoDeploy            bool                   `json:"nodeploy"`
	SkipRebuild         bool
	ImageID             string
	testPath            string
	Name                string
	NaclFileShasum      string
}

type testResponse struct {
	Result   bool    `json:"result"`   // Pass/Fail of the test
	Sent     int     `json:"sent"`     // Number of tests started
	Received int     `json:"received"` // Number of responses received
	Rate     float32 `json:"rate"`     // Requests pr second
	Raw      string  `json:"raw"`      // Raw output from test
}

type TestResult struct {
	Name              string        // Name of test
	Duration          time.Duration // Time to execute test
	SuccessPercentage float32       // Percentage success
	testResponse
}

type HostCommandTemplate struct {
	MothershipBinPathAndName string
	OriginalAlias            string
	ImageID                  string
	BuilderID                string
}

func (r TestResult) StringSlice() [][]string {
	var result string
	if r.Result {
		result = fmt.Sprintf("[%s]", aurora.BgGreen(" PASS "))
	} else {
		result = fmt.Sprintf("[%s]", aurora.BgRed(" FAIL "))
	}
	return [][]string{
		[]string{"Result", result},
		[]string{"Sent", strconv.Itoa(r.Sent)},
		[]string{"Received", strconv.Itoa(r.Received)},
		[]string{"Percentage", fmt.Sprintf("%.1f%%", r.SuccessPercentage)},
		[]string{"Rate", fmt.Sprintf("%.2f", r.Rate)},
		[]string{"Duration", r.Duration.Truncate(1 * time.Second).String()},
	}
}

func (c TestConfig) StringSlice() [][]string {
	var output [][]string
	if c.ClientCommandScript != "" {
		output = append(output, []string{"Path", c.ClientCommandScript})
		output = append(output, []string{"Script type", "[X] ClientCommandScript / [ ] HostCommandScript"})
	} else if c.HostCommandScript != "" {
		output = append(output, []string{"Path", c.HostCommandScript})
		output = append(output, []string{"Script type", "[ ] ClientCommandScript / [X] HostCommandScript"})
	}
	output = append(output, []string{"Custom service", c.CustomServicePath})
	if c.NoDeploy {
		output = append(output, []string{"No deploy", "[X]"})
	} else {
		output = append(output, []string{"No deploy", "[ ]"})
	}
	if c.SkipRebuild {
		output = append(output, []string{"Skip rebuild", "[X]"})
	} else {
		output = append(output, []string{"Skip rebuild", "[ ]"})
	}
	return output
}

// RunTest runs the clientCmdScript on either host or client1 level number of times and returns a TestResult
func (t *TestConfig) RunTest(iterations int, env environment.Environment, mother *mothership.Mothership) (TestResult, error) {
	// Prepare test before run
	logrus.Info("Preparing clients")
	if err := t.runScriptsOnClients(env, t.Setup); err != nil {
		return TestResult{}, fmt.Errorf("error preparing test: %v", err)
	}
	defer t.cleanupTest(mother, env)
	var results []TestResult
	logrus.Info("Starting test")
	for i := 0; i < iterations; i++ {
		var testOutput []byte
		var testResult TestResult
		testResult.Name = t.Name
		start := time.Now()
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
		testResult.Duration = time.Now().Sub(start)

		// Parse test results
		if err = json.Unmarshal(testOutput, &testResult); err != nil {
			return testResult, fmt.Errorf("could not parse testResults: %v", err)
		}

		// Calculate success
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
	end.Result = true // true until otherwise proven
	for _, result := range results {
		if !result.Result {
			end.Result = false
		}
		end.Name = result.Name
		end.Sent += result.Sent
		end.Received += result.Received
		end.Rate += result.Rate
		end.Duration += result.Duration
	}
	end.Rate = end.Rate / float32(len(results))
	end.SuccessPercentage = float32(end.Received) / float32(end.Sent) * 100
	return end
}
