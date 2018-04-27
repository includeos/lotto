package main

import (
	"github.com/mnordsletten/lotto/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	/*
		testConf := testFramework.TestConfig{
			ID:              "test1",
			TestEnvironment: "microLB-env",
			TestService:     "microLB",
			ClientCommand:   "<stress-test-microlb>",
			MonitorCommand:  "<monitor-cmd>",
		}
		// 1. Set up test environment
		env := &environment.Vcloud{}
		env.SetUp(testConf.TestEnvironment)

		// 2. Create starbase that connects to mothership
		starbase, err := mothership.CreateStarbase()
		if err != nil {
			logrus.Errorf("Error creating starbase: %v", err)
		}
		testConf.Starbase = starbase

		// Now everything is ready to start testing

		// 3. Build test service and liveupdate starbase
		uplink := "localhost:9090"
		if err := mothership.LaunchService(starbase, testConf.TestService, uplink); err != nil {
			logrus.Errorf("Error launching service: %v", err)
		}

		logrus.Debugf("Environment and service are now ready to begin")

		// 4. Start test towards target
		if err := testConf.TestLoop(env); err != nil {
			logrus.Errorf("Error starting test loop: %v", err)
		}
	*/
	cmd.RootCmd.Execute()

}
