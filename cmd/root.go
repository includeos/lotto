package cmd

import (
	"github.com/mnordsletten/lotto/environment"
	"github.com/mnordsletten/lotto/mothership"
	"github.com/mnordsletten/lotto/testFramework"
	"github.com/mnordsletten/lotto/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cmdEnv           string
	verboseLogging   bool
	setUpEnv         bool
	forceNewStarbase bool
	skipRebuildTest  bool
	skipVerifyEnv    bool
	numRuns          int
	loops            int

	mothershipConfigPath string
	envConfigPath        string
)

var RootCmd = &cobra.Command{
	Use:   "lotto TEST-FOLDER-PATH [TEST-FOLDER-PATH...]",
	Short: "Run tests by specifying test folders",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if verboseLogging {
			logrus.SetLevel(logrus.DebugLevel)
		}
		// Environment creation
		env, err := envFromConfig(envConfigPath, cmdEnv)
		if err != nil {
			logrus.Fatalf("Could not set up environment: %v", err)
		}
		if setUpEnv {
			err = env.Create()
			if err != nil {
				logrus.Fatalf("Error setting up environment: %v", err)
			}
		}
		if !skipVerifyEnv {
			logrus.Info("Verifying environment")
			if err = environment.VerifyEnv(env); err != nil {
				logrus.Fatalf("Error verifying env: %v", err)
			}
		}

		// Mothership setup
		var mother *mothership.Mothership
		mother, err = mothershipFromConfig(mothershipConfigPath, env)
		if err != nil {
			logrus.Fatalf("Could not set up Mothership: %v", err)
		}

		// Only create a new starbase if requested, or there is no connected starbase to use
		if forceNewStarbase || !mother.CheckStarbaseIDInUse() {
			logrus.Infof("Launching a new clean Starbase")
			// Push nacl, build and download clean starbase image
			if err = mother.LaunchCleanStarbase(env); err != nil {
				logrus.Fatalf("error creating clean starbase: %v", err)
			}
		}

		// Test setup
		tests := make([]*testFramework.TestConfig, len(args))
		for i, arg := range args {
			tests[i], err = testFramework.ReadFromDisk(arg)
			if err != nil {
				logrus.Fatalf("Could not read test spec: %v", err)
			}
		}
		for loopIndex := 0; loopIndex < loops || loops == 0; loopIndex++ {
			logrus.Infof("Test loop nr: %d, numRuns: %d", loopIndex+1, numRuns)
			for _, test := range tests {
				// Boot NaCl service to starbase
				if !skipRebuildTest {
					if err = mother.DeployNacl(test.NaclFile); err != nil {
						logrus.Fatalf("Could not deploy: %v", err)
					}
				}
				// Run client command
				result, err := test.RunTest(numRuns, env)
				if err != nil {
					logrus.Warningf("error running test %v", err)
				}
				logrus.Info(result)
				util.StructToCsvOutput(result, "testResults")
				health := mother.CheckInstanceHealth()
				logrus.Info(health)
				util.StructToCsvOutput(health, "instanceHealth")
			}
		}
	},
}

func init() {
	RootCmd.Flags().StringVar(&cmdEnv, "env", "fusion", "environment to use")
	RootCmd.Flags().BoolVarP(&verboseLogging, "verbose", "v", false, "verobse output")
	RootCmd.Flags().BoolVar(&setUpEnv, "create-env", false, "set up environment")
	RootCmd.Flags().BoolVar(&forceNewStarbase, "force-new-starbase", false, "create a new starbase")
	RootCmd.Flags().BoolVar(&skipRebuildTest, "skipRebuildTest", false, "push new nacl and rebuild before deploying")
	RootCmd.Flags().BoolVar(&skipVerifyEnv, "skipVerifyEnv", false, "skip environment verification")
	RootCmd.Flags().IntVarP(&numRuns, "numTestRuns", "n", 1, "number of test iterations to run for each test, 0 means infinite")
	RootCmd.Flags().IntVarP(&loops, "loops", "l", 1, "number of loops for all tests to run, 0 means infinite")

	RootCmd.Flags().StringVar(&mothershipConfigPath, "mship-config", "config-mothership.json", "Mothership config file")
	RootCmd.Flags().StringVar(&envConfigPath, "env-config", "config-environment.json", "Environments config file")
}
