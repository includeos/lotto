package cmd

import (
	"github.com/mnordsletten/lotto/environment"
	"github.com/mnordsletten/lotto/mothership"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	builderName      string
	cmdEnv           string
	tag              string
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
	Use:   "lotto <test-folder-path> [test-folder-path...]",
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
		// Prepare builder
		if err = mother.PrepareBuilder(builderName); err != nil {
			logrus.Fatalf("Could not prepare builder with name: %s: %v", builderName, err)
		}
		logrus.Infof("Prepared builder with name: %s on Mothership", builderName)

		// Only create a new starbase if requested, or there is no connected starbase to use
		if forceNewStarbase || !mother.CheckStarbaseIDInUse() {
			logrus.Infof("Launching a new clean Starbase")
			// Push nacl, build and download clean starbase image
			if err = mother.LaunchCleanStarbase(env); err != nil {
				logrus.Fatalf("error creating clean starbase: %v", err)
			}
		}

		// Test setup
		tests, err := getTestsToRun(args)
		// Run the tests
		for loopIndex := 0; loopIndex < loops || loops == 0; loopIndex++ {
			logrus.Infof("Test loop nr: %d, numRuns: %d", loopIndex+1, numRuns)
			for _, test := range tests {
				if skipRebuildTest {
					test.SkipRebuild = true
				}
				if err = testProcedure(test, env, mother); err != nil {
					logrus.Warningf("error running test %s: %v", test.Name, err)
				}
			}
		}
	},
}

func init() {
	RootCmd.Flags().StringVarP(&builderName, "buildername", "b", "", "The IncludeOS builder to use for running the test")
	RootCmd.MarkFlagRequired("buildername")
	RootCmd.Flags().StringVar(&cmdEnv, "env", "fusion", "environment to use")
	RootCmd.Flags().BoolVarP(&verboseLogging, "verbose", "v", false, "verobse output")
	RootCmd.Flags().BoolVar(&setUpEnv, "create-env", false, "set up environment")
	RootCmd.Flags().BoolVar(&forceNewStarbase, "force-new-starbase", false, "create a new starbase")
	RootCmd.Flags().BoolVar(&skipRebuildTest, "skipRebuildTest", false, "push new nacl and rebuild before deploying")
	RootCmd.Flags().BoolVar(&skipVerifyEnv, "skipVerifyEnv", false, "skip environment verification")
	RootCmd.Flags().IntVarP(&numRuns, "numTestRuns", "n", 1, "number of test iterations to run for each test")
	RootCmd.Flags().IntVarP(&loops, "loops", "l", 1, "number of loops for all tests to run, 0 means infinite")
	RootCmd.Flags().StringVarP(&tag, "tag", "t", "", "Tag to give folder that stores testResults, if none then testResults are not saved")

	RootCmd.Flags().StringVar(&mothershipConfigPath, "mship-config", "config-mothership.json", "Mothership config file")
	RootCmd.Flags().StringVar(&envConfigPath, "env-config", "config-environment.json", "Environments config file")
}
