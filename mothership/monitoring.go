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

	// Check for any crashes, and get the latest crash output if necessary
	crashIDs, err := m.getAllCrashesArray()
	if err != nil {
		logrus.Warningf("could not get crashes array: %v", err)
		return i
	}
	logrus.Debugf("All crashes: %v", crashIDs)
	if newCrash := m.getNewestCrashID(crashIDs); newCrash != "" {
		i.PanicContent, err = m.getSingleCrashOutput(newCrash)
		if err != nil {
			logrus.Warningf("could not get crash output: %v", err)
			return i
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

// getNewestCrashID returns the crashID of the crashes that are new
func (m *Mothership) getNewestCrashID(crashIDs []string) string {
	defer func() {
		m.lastCheckTime = time.Now()
	}()
	latestCrashID := crashIDs[len(crashIDs)-1]
	latestCrash := strings.TrimLeft(latestCrashID, "panic_")
	latestCrash = strings.TrimRight(latestCrash, ".txt")
	latestCrashTime, err := time.Parse(time.RFC3339, latestCrash)
	if err != nil {
		logrus.Warningf("could not parse panic time %s: %v", latestCrash, err)
		return ""
	}
	logrus.Debugf("latest crashID: %s", latestCrash)
	if latestCrashTime.Before(m.lastCheckTime) {
		return ""
	}
	return latestCrashID
}

// returns the full list of the names of all crash reports
func (m *Mothership) getAllCrashesArray() ([]string, error) {
	type crashes []string
	var c crashes
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

func (m *Mothership) getSingleCrashOutput(crashID string) (string, error) {
	request := fmt.Sprintf("instance-crash %s %s", m.alias, crashID)
	output, err := m.bin(request)
	if err != nil {
		return "", err
	}
	return output, nil
}
