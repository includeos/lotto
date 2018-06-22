package testFramework

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/mnordsletten/lotto/environment"
	"github.com/sirupsen/logrus"
)

type TestConfig struct {
	ID                  string                 `json:"id"`
	NaclFile            string                 `json:"naclfile"`
	ClientCommandScript string                 `json:"clientcommandscript"`
	Setup               environment.SSHClients `json:"setup"`
	Level1              int                    `json:"level1"`
	Level2              int                    `json:"level2"`
	Level3              int                    `json:"level3"`
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
}

func (tr TestResult) String() string {
	return fmt.Sprintf("Name: %s, Sent: %d, Received: %d Percentage: %.1f%%", tr.Name, tr.Sent, tr.Received, tr.SuccessPercentage)
}

// RunTest runs the clientCmdScript on client1 level number of times and returns a TestResult
func (t *TestConfig) RunTest(level int, env environment.Environment) TestResult {
	if err := t.prepareTest(env); err != nil {
		logrus.Fatalf("error preparing test: %v", err)
	}
	var results []TestResult
	for i := 0; i < level; i++ {
		testOutput, err := env.RunClientCmdScript(1, t.ClientCommandScript)
		if err != nil {
			logrus.Fatalf("could not run client command script: %v", err)
			os.Exit(1)
		}
		var testResult TestResult
		testResult.Time = time.Now().Format(time.RFC3339)
		testResult.Name = path.Base(t.testPath)
		if err = json.Unmarshal(testOutput, &testResult); err != nil {
			logrus.Fatalf("could not parse testResults: %v", err)
		}
		testResult.SuccessPercentage = float32(testResult.Received) / float32(testResult.Sent) * 100
		results = append(results, testResult)
	}
	return combineTestResults(results)
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
		end.SuccessPercentage = float32(end.Received) / float32(end.Sent) * 100
	}
	end.Time = time.Now().Format(time.RFC3339)
	return end
}
