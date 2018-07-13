package mothership

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type InstanceHealth struct {
	Status       string `json:"status"`
	TotalPanics  int    `json:"panics"`
	IosVersion   string `json:"version"`
	Time         string
	PanicContent string
}

func (i InstanceHealth) String() string {
	return fmt.Sprintf("Status: %s, IOSVersion: %s, %s", i.Status, i.IosVersion, i.PanicContent)
}

func (m *Mothership) CheckInstanceHealth() InstanceHealth {
	i, err := m.getInstanceInfo(m.alias)
	if err != nil {
		logrus.Warningf("Could not get instance info: %v", err)
	}

	// Check for any panics, and get the latest panic output if necessary
	panicIDs, err := m.getAllPanicsArray()
	if err != nil {
		logrus.Warningf("could not get panics array: %v", err)
		return i
	}
	logrus.Debugf("All panics: %v", panicIDs)
	if len(panicIDs) > 0 {
		if newPanic := m.getNewestPanicID(panicIDs); newPanic != "" {
			i.PanicContent, err = m.getSinglePanicOutput(newPanic)
			if err != nil {
				logrus.Warningf("could not get panic output: %v", err)
				return i
			}
		}
	}
	return i
}

func (m *Mothership) getInstanceInfo(id string) (InstanceHealth, error) {
	var i InstanceHealth
	i.Time = time.Now().Format(time.RFC3339)

	request := fmt.Sprintf("inspect-instance %s -o json", id)
	output, err := m.bin(request)
	if err != nil {
		return i, err
	}
	if err := json.Unmarshal([]byte(output), &i); err != nil {
		return i, fmt.Errorf("error unmarshaling instanceInfo: %v", err)
	}

	return i, nil
}

// getNewestPanicID returns the panicID of the panics that are new
func (m *Mothership) getNewestPanicID(panicIDs []string) string {
	defer func() {
		m.lastCheckTime = time.Now()
	}()
	latestPanicID := panicIDs[len(panicIDs)-1]
	latestPanic := strings.TrimLeft(latestPanicID, "panic_")
	latestPanic = strings.TrimRight(latestPanic, ".txt")
	latestPanicTime, err := time.Parse(time.RFC3339, latestPanic)
	if err != nil {
		logrus.Warningf("could not parse panic time %s: %v", latestPanic, err)
		return ""
	}
	logrus.Debugf("latest panicID: %s", latestPanic)
	if latestPanicTime.Before(m.lastCheckTime) {
		return ""
	}
	return latestPanicID
}

// returns the full list of the names of all panic reports
func (m *Mothership) getAllPanicsArray() ([]string, error) {
	type panics []string
	var c panics
	request := fmt.Sprintf("instance-crashes %s -o json", m.alias)
	output, err := m.bin(request)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(output), &c); err != nil {
		return nil, fmt.Errorf("error unmarshaling instance-crashes: %v", err)
	}
	return c, nil
}

func (m *Mothership) getSinglePanicOutput(panicID string) (string, error) {
	request := fmt.Sprintf("instance-crash %s %s", m.alias, panicID)
	output, err := m.bin(request)
	if err != nil {
		return "", err
	}
	return output, nil
}
