package cmd

import (
	"fmt"
	"strings"

	"github.com/mnordsletten/lotto/environment"
	"github.com/mnordsletten/lotto/mothership"
	"github.com/mnordsletten/lotto/prettyoutput"
	"github.com/mnordsletten/lotto/reporting"
	"github.com/mnordsletten/lotto/testFramework"
	"github.com/sirupsen/logrus"
)

func testProcedure(test *testFramework.TestConfig, env environment.Environment, mother *mothership.Mothership) (bool, error) {
	pretty := pretty.NewPrettyTest(test.Name)
	pretty.PrintHeader()
	pretty.PrintTable(test.StringSlice())

	// BUILD & DEPLOY. 3 options:
	// 1. Push NaCl and build on Mothership
	// 2. Build service locally using docker
	// 3. No building at all
	if err := build(test, mother); err != nil {
		return false, fmt.Errorf("error building: %v", err)
	}

	// TEST script. 2 options:
	// 1. HOSTcommandscript
	// 2. CLIENTcommandscript
	// Run client command
	// numRuns flag taken into account
	result, err := test.RunTest(numRuns, env, mother)
	if err != nil {
		return false, fmt.Errorf("error running test %v", err)
	}

	// RESULTS print test results
	pretty.PrintTable(result.StringSlice())

	// VERIFY starbase status
	health := mother.CheckInstanceHealth()
	logrus.Info(health)

	pretty.EndTest()
	reporting.SendReport(reporting.Dashboard{
		Address:           "http://localhost:7070/upload",
		MothershipVersion: "v1",
		IncludeOSVersion:  builderName,
		Environment:       cmdEnv,
		TestResult:        result,
	})
	return result.Success, nil
}

func getTestsToRun(possibleTests []string) ([]*testFramework.TestConfig, error) {
	// Filter out the tests that should be skipped (folder name starts with "skip")
	var testsToRun []string
	for _, arg := range possibleTests {
		// Skipping tests starting with "tests/skip"
		if !strings.HasPrefix(arg, "tests/skip") {
			testsToRun = append(testsToRun, arg)
		} else {
			logrus.Warningf("Skipping test %s", arg)
		}
	}
	// Get the TestConfig for every test that should be run
	var tests []*testFramework.TestConfig
	for _, testPath := range testsToRun {
		test, err := testFramework.ReadFromDisk(testPath)
		if err != nil {
			return nil, fmt.Errorf("Could not read test spec: %v", err)
		}
		tests = append(tests, test)
	}
	return tests, nil
}

func build(test *testFramework.TestConfig, mother *mothership.Mothership) error {
	// Return early if skip rebuild has been set
	if test.SkipRebuild {
		return nil
	}
	var err error
	// Boot NaCl service to starbase, only if NaclFile is specified
	if test.NaclFile != "" {
		if test.NaclFileShasum, test.ImageID, err = mother.DeployNacl(test.NaclFile); err != nil {
			return fmt.Errorf("could not deploy: %v", err)
		}
	}
	// Build and deploy custom service if specified
	if test.CustomServicePath != "" {
		if test.ImageID, err = mother.BuildPushAndDeployCustomService(test.CustomServicePath, builderName, test.NoDeploy); err != nil {
			return fmt.Errorf("could not build and push custom service: %v", err)
		}
	}
	return nil
}

/*
func saveResultsToDisk(result testFramework.TestResult, mother *mothership.Mothership) error {
	if len(tag) > 0 // blah blah
	// Create folder with name of versions getting tested
	mVersion, err := mother.ServerVersion()
	if err != nil {
		return fmt.Errorf("error getting mothership server version: %v", err)
	}
	iosVersion, err := mother.StarbaseVersion()
	if err != nil {
		return fmt.Errorf("error getting starbase IncludeOS version: %v", err)
	}
	folderPath := path.Join("testResults", fmt.Sprintf("mothership.%s_IncludeOS.%s_%s", mVersion, iosVersion, tag))
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		return fmt.Errorf("Could not create testResults folder: %v", err)
	}
	if len(result.Name) > 0 {
		util.StructToCsvOutput(result, path.Join(folderPath, result.Name))
	}
	healthName := fmt.Sprintf("instanceHealth-%s", time.Now().Format("2006-01-02"))
	util.StructToCsvOutput(health, path.Join(folderPath, healthName))
}
*/
