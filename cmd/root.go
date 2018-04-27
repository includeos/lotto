package cmd

import (
	"github.com/mnordsletten/lotto/environment"
	"github.com/mnordsletten/lotto/mothership"
	"github.com/mnordsletten/lotto/testFramework"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cmdEnv string
)

var RootCmd = &cobra.Command{
	Use: "lotto",
	Run: func(cmd *cobra.Command, args []string) {
		envSettings := environment.EnvSettings{
			Username: "User",
			Password: "Password",
			AppName:  "TestApp",
			Address:  "remote.org",
		}
		// Create test definition
		testSetup := &testFramework.TestConfig{
			Nacl:          "microLB",
			ClientCommand: "bash ping a lot",
		}

		// Environment setup
		env, err := newEnvironment(cmdEnv, envSettings)
		if err != nil {
			logrus.Fatalf("Could not set up environment: %v", err)
		}
		testSetup.TestEnvironment = env
		testSetup.TestEnvironment.Create()
		_, testSetup.Starbase = testSetup.TestEnvironment.BootStarbase()

		// Set up Mothership
		testSetup.Mothership = mothership.NewMothership("10.0.0.1", "10.0.0.10", "martin", "123")

		// Boot NaCl service to starbase
		testSetup.Mothership.DeployNacl(testSetup.Nacl, testSetup.Starbase)

		// Run client command
		testOutput, _ := testSetup.TestEnvironment.RunClientCmd(testSetup.ClientCommand)

		// Set up monitoring ???

		// Parse output from command
		logrus.Infof("Test output: %s", testOutput)

	},
}

func init() {
	RootCmd.Flags().StringVar(&cmdEnv, "env", "vcloud", "environment to use")
}
