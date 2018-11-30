package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/mnordsletten/lotto/environment"
	"github.com/mnordsletten/lotto/mothership"
	"github.com/mnordsletten/lotto/testFramework"
	"github.com/mnordsletten/lotto/util"
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
		// Filter out the tests that should be skipped (folder name starts with "skip")
		var testsToRun []string
		for _, arg := range args {
			// Skipping tests starting with "tests/skip"
			if !strings.HasPrefix(arg, "tests/skip") {
				testsToRun = append(testsToRun, arg)
			} else {
				logrus.Warningf("Skipping test %s", arg)
			}
		}
		// Get the TestConfig for every test that should be run
		tests := make([]*testFramework.TestConfig, len(testsToRun))
		for i, testPath := range testsToRun {
			tests[i], err = testFramework.ReadFromDisk(testPath)
			if err != nil {
				logrus.Fatalf("Could not read test spec: %v", err)
			}
		}
		// Run the tests
		var testFailed bool
		// loops flag taken into account
		for loopIndex := 0; loopIndex < loops || loops == 0; loopIndex++ {
			logrus.Infof("Test loop nr: %d, numRuns: %d", loopIndex+1, numRuns)
			for _, test := range tests {
				// Boot NaCl service to starbase, only if NaclFile is specified
				if !skipRebuildTest {
					if test.NaclFile != "" {
						if test.NaclFileShasum, test.ImageID, err = mother.DeployNacl(test.NaclFile); err != nil {
							logrus.Warningf("Could not deploy: %v", err)
							testFailed = true
							continue
						}
					}
				}
				// Build and deploy custom service if specified
				if test.CustomServicePath != "" {
					if test.ImageID, err = mother.BuildPushAndDeployCustomService(test.CustomServicePath, builderName, test.Deploy); err != nil {
						testFailed = true
						logrus.Warningf("could not build and push custom service: %v", err)
					}
				}
				// Run client command
				// numRuns flag taken into account
				result, err := test.RunTest(numRuns, env, mother)
				if err != nil {
					testFailed = true
					logrus.Warningf("error running test %v", err)
				}
				// Process results
				logrus.Info(result)
				health := mother.CheckInstanceHealth()
				logrus.Info(health)
				if len(tag) > 0 {
					// Create folder with name of versions getting tested
					mVersion, err := mother.ServerVersion()
					if err != nil {
						logrus.Warningf("error getting mothership server version: %v", err)
					}
					iosVersion, err := mother.StarbaseVersion()
					if err != nil {
						logrus.Warningf("error getting starbase IncludeOS version: %v", err)
					}
					folderPath := path.Join("testResults", fmt.Sprintf("mothership.%s_IncludeOS.%s_%s", mVersion, iosVersion, tag))
					if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
						logrus.Fatalf("Could not create testResults folder: %v", err)
					}
					if len(result.Name) > 0 {
						util.StructToCsvOutput(result, path.Join(folderPath, test.Name))
					}
					healthName := fmt.Sprintf("instanceHealth-%s", time.Now().Format("2006-01-02"))
					util.StructToCsvOutput(health, path.Join(folderPath, healthName))
				}
			}
		}

		if testFailed {
			logrus.Fatal("A test has failed")
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
