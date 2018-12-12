package reporting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mnordsletten/lotto/testFramework"
)

type Dashboard struct {
	Address string
	testFramework.TestResult
	MothershipVersion string `json:"mothershipVersion"`
	IncludeOSVersion  string `json:"includeosVersion"`
	Environment       string `json:"environment"`
}

func SendReport(config Dashboard) error {
	fmt.Println("Sending report with contents: ", config)

	jsonStr, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error encoding JSON: %v", err)
	}

	req, err := http.NewRequest("POST", config.Address, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending report to dashboard: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error, did not get statusOK from dashboard when uploading test result")
	}
	return nil
}
