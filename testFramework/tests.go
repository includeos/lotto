package testFramework

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mnordsletten/lotto/environment"
	"github.com/sirupsen/logrus"
)

type TestConfig struct {
	ID                  string `json:"id"`
	NaclFile            string `json:"naclfile"`
	ClientCommandScript string `json:"clientcommandscript"`
	Level1              int    `json:"level1"`
	Level2              int    `json:"level2"`
	Level3              int    `json:"level3"`
}

type TestResult struct {
	Sent              int     // Total number of requests sent
	Received          int     // Total number of replies received
	Rate              float32 // Requests pr second
	Avg               float32 // Average response time
	SuccessPercentage float32 // Percentage of packets that pass
	Raw               string  // Raw output from the command
}

func (tr TestResult) String() string {
	return fmt.Sprintf("Sent: %d, Received: %d Percentage: %.1f%%", tr.Sent, tr.Received, tr.SuccessPercentage)
}

func (t *TestConfig) RunTest(level int, env environment.Environment) TestResult {
	var results []TestResult
	for i := 0; i < level; i++ {
		testOutput, err := env.RunClientCmdScript(t.ClientCommandScript)
		if err != nil {
			logrus.Fatalf("could not run client command script: %v", err)
			os.Exit(1)
		}
		var testResult TestResult
		if err = json.Unmarshal(testOutput, &testResult); err != nil {
			logrus.Fatalf("could not parse testResults: %v", err)
		}
		testResult.SuccessPercentage = float32(testResult.Received) / float32(testResult.Sent) * 100
		logrus.Infof("%s", testResult)
		results = append(results, testResult)
	}
	return combineTestResults(results)
}

func combineTestResults(results []TestResult) TestResult {
	end := TestResult{}
	for _, result := range results {
		end.Sent += result.Sent
		end.Received += result.Received
		end.Rate += result.Rate
		end.SuccessPercentage = float32(end.Received) / float32(end.Sent) * 100
	}
	return end
}
