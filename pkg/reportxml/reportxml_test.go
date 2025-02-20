package reportxml

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	testCases := []struct {
		testSettings  *settings
		reportCaseTag string
		IDTag         string
		ParameterTag  string
	}{
		{
			testSettings: &settings{
				ParameterTag: "parameter",
				CaseTag:      "testcase-id",
				IDTag:        "test_id",
			},
		},
		{
			reportCaseTag: "test1",
			IDTag:         "test_id",
			ParameterTag:  "test3",
			testSettings: &settings{
				ParameterTag: "test3",
				CaseTag:      "test1",
				IDTag:        "test_id",
			},
		},
	}

	for _, testCase := range testCases {
		if testCase.reportCaseTag != "" {
			t.Setenv("REPORT_CASE_TAG", testCase.reportCaseTag)
			t.Setenv("REPORT_PARAMETER_TAG", testCase.ParameterTag)
		}

		setting, err := newConfig()
		assert.Nil(t, err)
		assert.NotNil(t, setting)
		assert.Equal(t, testCase.testSettings, setting)
	}
}

func TestSetDefaultTag(t *testing.T) {
	testCases := []struct {
		testSettings *settings
	}{
		{
			testSettings: &settings{
				ParameterTag: "parameter",
				CaseTag:      "testcase-id",
				IDTag:        "test_id",
			},
		},
	}

	for _, testCase := range testCases {
		setting, err := newConfig()
		assert.Nil(t, err)
		setting.setDefaultTag()
		assert.Nil(t, err)
		assert.NotNil(t, setting)
		assert.Equal(t, testCase.testSettings, setting)
	}
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		destFile   string
		projectTag string
		fileExist  bool
	}{
		{
			destFile:   "test",
			projectTag: "test2",
			fileExist:  true,
		},
		{
			destFile:   "",
			projectTag: "test2",
			fileExist:  false,
		},
		{
			destFile:   "",
			projectTag: "",
			fileExist:  false,
		},
	}

	for _, testCase := range testCases {
		testReport := ginkgo.Report{}
		Create(testReport, testCase.destFile, testCase.projectTag)

		if testCase.fileExist {
			assert.FileExists(t, testCase.destFile)
			assert.Nil(t, os.RemoveAll(testCase.destFile))
		} else {
			assert.NoFileExists(t, testCase.destFile)
		}
	}
}

func TestID(t *testing.T) {
	testCases := []struct {
		testTag string
	}{
		{
			testTag: "test",
		},
		{
			testTag: "",
		},
	}
	for _, testCase := range testCases {
		tag := ID(testCase.testTag)
		assert.Contains(t, tag, testCase.testTag)
	}
}

func TestSetProperty(t *testing.T) {
	testCases := []struct {
		propertyKey   string
		propertyValue string
	}{
		{
			propertyKey:   "test",
			propertyValue: "test2",
		},
		{
			propertyKey:   "test3",
			propertyValue: "test4",
		},
	}
	for _, testCase := range testCases {
		property := SetProperty(testCase.propertyKey, testCase.propertyValue)
		assert.Contains(t, property, fmt.Sprintf("parameter-%s:%s", testCase.propertyKey, testCase.propertyValue))
	}
}

func TestSetTestID(t *testing.T) {
	testCases := []struct {
		projectTag string
		labels     [][]string
		valid      bool
	}{
		{
			labels:     [][]string{{"test_id:1111"}},
			valid:      true,
			projectTag: "AAAA",
		},
		{
			labels:     [][]string{{"1111"}},
			valid:      false,
			projectTag: "BBBB",
		},
		{
			labels:     [][]string{},
			valid:      false,
			projectTag: "BBBB",
		},
		{
			labels:     [][]string{{"test_id:1111"}},
			valid:      true,
			projectTag: "",
		},
	}

	for _, testCase := range testCases {
		report := ginkgo.SpecReport{
			ContainerHierarchyLabels: testCase.labels,
		}
		testID := setTestID(report, testCase.projectTag)

		if testCase.valid {
			assert.NotNil(t, testID)
			assert.Equal(t, testID.Name, "testcase-id")
			assert.Contains(t, testID.Value, testCase.projectTag)
		} else {
			assert.Nil(t, testID)
		}
	}
}

func TestReportSetProperty(t *testing.T) {
	testCases := []struct {
		labels [][]string
		valid  bool
	}{
		{
			labels: [][]string{{"parameter-id:1111"}},
			valid:  true,
		},
		{
			labels: [][]string{{"id:1111"}},
			valid:  false,
		},
		{
			labels: [][]string{{"1111"}},
			valid:  false,
		},
	}
	for _, testCase := range testCases {
		report := ginkgo.SpecReport{
			ContainerHierarchyLabels: testCase.labels}

		properties := setProperty(report)
		if testCase.valid {
			assert.NotNil(t, properties)
		} else {
			assert.Nil(t, properties)
		}
	}
}

func TestSetFailureMessage(t *testing.T) {
	testCases := []struct {
		state       types.SpecState
		failMessage string
	}{
		{
			state:       types.SpecStateFailed,
			failMessage: "Fail test",
		},
		{
			state: types.SpecStatePassed,
		},
	}
	for _, testCase := range testCases {
		report := ginkgo.SpecReport{
			Failure: types.Failure{
				Message: testCase.failMessage,
			},
			State: testCase.state,
		}

		failMessage := setFailureMessage(report)

		if testCase.failMessage != "" {
			assert.Contains(t, failMessage.Message, testCase.failMessage)
		} else {
			assert.Nil(t, failMessage)
		}
	}
}

func TestSetSkipMessage(t *testing.T) {
	testCases := []struct {
		state       types.SpecState
		failMessage string
	}{
		{
			state:       types.SpecStateSkipped,
			failMessage: "Skip test",
		},
		{
			state: types.SpecStatePassed,
		},
	}
	for _, testCase := range testCases {
		report := ginkgo.SpecReport{
			Failure: types.Failure{
				Message: testCase.failMessage,
			},
			State: testCase.state,
		}

		failMessage := setSkipMessage(report)

		if testCase.failMessage != "" {
			assert.Contains(t, failMessage.Message, testCase.failMessage)
		} else {
			assert.Nil(t, failMessage)
		}
	}
}

func TestSetTestSuite(t *testing.T) {
	testCases := []struct {
		report      types.Report
		failMessage string
	}{
		{
			report: ginkgo.Report{
				SuiteDescription: "test",
				RunTime:          10 * time.Minute,
				SpecReports:      types.SpecReports{},
			},
		},
	}
	for _, testCase := range testCases {
		testSuite := setTestSuite(testCase.report)
		assert.NotNil(t, testSuite)
	}
}

func TestCreateNewReportFile(t *testing.T) {
	testCases := []struct {
		testSuite   *TestSuite
		failMessage string
	}{
		{
			testSuite: &TestSuite{},
		},
	}
	for _, testCase := range testCases {
		createNewReportFile("test", testCase.testSuite)
		assert.FileExists(t, "test")
		assert.Nil(t, os.RemoveAll("test"))
	}
}

func TestAppendToExistingReportFile(t *testing.T) {
	testCases := []struct {
		testSuite *TestSuite
		panic     bool
		dstFile   string
	}{
		{
			panic:     true,
			dstFile:   "",
			testSuite: generateTestSuite(),
		},
		{
			panic:     false,
			dstFile:   "test",
			testSuite: generateTestSuite(),
		},
	}
	for _, testCase := range testCases {
		createNewReportFile("test", &TestSuite{})
		assert.FileExists(t, "test")

		if testCase.panic {
			assert.Panics(t, func() { appendToExistingReportFile(testCase.dstFile, testCase.testSuite) })
		} else {
			appendToExistingReportFile(testCase.dstFile, testCase.testSuite)
		}

		assert.FileExists(t, "test")
		assert.Nil(t, os.RemoveAll("test"))
	}
}

func TestGenerateReportXMLFile(t *testing.T) {
	testCases := []struct {
		testSuite *TestSuite
		exist     bool
		dstFile   string
	}{
		{
			exist:     true,
			dstFile:   "test",
			testSuite: generateTestSuite(),
		},
		{
			exist:     false,
			dstFile:   "test",
			testSuite: generateTestSuite(),
		},
	}
	for _, testCase := range testCases {
		if testCase.exist {
			createNewReportFile(testCase.dstFile, &TestSuite{})
			assert.FileExists(t, "test")
		} else {
			assert.NoFileExists(t, "test")
		}

		generateReportXMLFile(testCase.dstFile, testCase.testSuite)
		assert.FileExists(t, "test")
		assert.Nil(t, os.RemoveAll("test"))
	}
}

func TestFailureTypeForState(t *testing.T) {
	testCases := []struct {
		state          types.SpecState
		expectedOutput string
	}{
		{
			state:          types.SpecStateFailed,
			expectedOutput: "Failure",
		},
		{
			state:          types.SpecStateInterrupted,
			expectedOutput: "Interrupted",
		},
		{
			state:          types.SpecStatePanicked,
			expectedOutput: "Panic",
		},
	}
	for _, testCase := range testCases {
		failureType := failureTypeForState(testCase.state)
		assert.Equal(t, failureType, testCase.expectedOutput)
	}
}

func TestFailureMessage(t *testing.T) {
	testCases := []struct {
		failure types.Failure
	}{
		{
			failure: types.Failure{
				Message: "test failure",
				FailureNodeLocation: types.CodeLocation{
					FileName:       "testfile",
					LineNumber:     10,
					FullStackTrace: "trace",
					CustomMessage:  "debug",
				},
				Location: types.CodeLocation{
					FileName:       "testfile",
					LineNumber:     10,
					FullStackTrace: "trace",
					CustomMessage:  "debug",
				},
			},
		},
	}
	for _, testCase := range testCases {
		failureType := failureMessage(testCase.failure)
		assert.Equal(t, fmt.Sprintf(
			"%s\n%s\n%s",
			testCase.failure.FailureNodeLocation.String(), testCase.failure.Message, testCase.failure.Location.String()),
			failureType)
	}
}

func generateTestSuite() *TestSuite {
	return &TestSuite{
		Name:     "new test suite",
		Tests:    1,
		Skipped:  0,
		Failures: 0,
		Time:     300,
		TestCases: []TestCase{{
			Name:      "test1",
			SystemOut: "allSet",
		},
		},
	}
}
