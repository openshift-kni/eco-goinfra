package kmm

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestNewModLoaderContainerBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		expectedError string
	}{
		{
			name:          "kmod",
			expectedError: "",
		},
		{
			name:          "",
			expectedError: "'modName' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testModuleLoaderContainerBuilder := NewModLoaderContainerBuilder(testCase.name)
		assert.Equal(t, testCase.expectedError, testModuleLoaderContainerBuilder.errorMsg)
		assert.NotNil(t, testModuleLoaderContainerBuilder.definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testModuleLoaderContainerBuilder.definition.Modprobe.ModuleName)
		}
	}
}

func TestModuleLoaderContainerWithModprobeSpec(t *testing.T) {
	testCases := []struct {
		dirName            string
		fwPath             string
		parameters         []string
		args               []string
		rawargs            []string
		moduleLoadingOrder []string
	}{
		{
			dirName:            "",
			fwPath:             "",
			parameters:         nil,
			moduleLoadingOrder: nil,
			args:               nil,
			rawargs:            nil,
		},
		{
			dirName:            "test",
			fwPath:             "test",
			parameters:         []string{"one", "two"},
			moduleLoadingOrder: []string{"one", "two"},
			args:               []string{},
			rawargs:            []string{},
		},
		{
			dirName:            "test",
			fwPath:             "test",
			parameters:         []string{"one", "two"},
			moduleLoadingOrder: []string{"one", "two"},
			args:               []string{"arg"},
			rawargs:            []string{},
		},
		{
			dirName:            "test",
			fwPath:             "test",
			parameters:         []string{"one", "two"},
			moduleLoadingOrder: []string{"one", "two"},
			args:               []string{},
			rawargs:            []string{"arg"},
		},
		{
			dirName:            "test",
			fwPath:             "test",
			parameters:         []string{"one", "two"},
			moduleLoadingOrder: []string{"one", "two"},
			args:               []string{"arg"},
			rawargs:            []string{"rawarg1", "rawargs2"},
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewModLoaderContainerBuilder("test")
		testBuilder.WithModprobeSpec(testCase.dirName, testCase.fwPath,
			testCase.parameters, testCase.args, testCase.rawargs, testCase.moduleLoadingOrder)

		assert.Equal(t, testCase.dirName, testBuilder.definition.Modprobe.DirName)
		assert.Equal(t, testCase.fwPath, testBuilder.definition.Modprobe.FirmwarePath)
		assert.Equal(t, testCase.parameters, testBuilder.definition.Modprobe.Parameters)
		assert.Equal(t, testCase.moduleLoadingOrder, testBuilder.definition.Modprobe.ModulesLoadingOrder)

		if len(testCase.args) > 0 {
			assert.Equal(t, testCase.args, testBuilder.definition.Modprobe.Args.Load)
		}

		if len(testCase.rawargs) > 0 {
			assert.Equal(t, testCase.rawargs, testBuilder.definition.Modprobe.RawArgs.Load)
		}
	}
}

func TestModuleLoaderContainerWithImagePullPolicy(t *testing.T) {
	testCases := []struct {
		imagePolicy   string
		expectedError string
	}{
		{
			imagePolicy:   "",
			expectedError: "'policy' can not be empty",
		},
		{
			imagePolicy:   "SomePolicy",
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewModLoaderContainerBuilder("test")
		testBuilder.WithImagePullPolicy(testCase.imagePolicy)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, corev1.PullPolicy(testCase.imagePolicy), testBuilder.definition.ImagePullPolicy)
		}
	}
}

func TestModuleLoaderContainerWithKernelMapping(t *testing.T) {
	testCases := []struct {
		mapping       *v1beta1.KernelMapping
		expectedError string
	}{
		{
			mapping:       buildRegExKernelMapping(""),
			expectedError: "'mapping' can not be empty nil",
		},
		{
			mapping:       buildRegExKernelMapping("^.+$"),
			expectedError: "",
		},
		{
			mapping:       buildLiteralKernelMapping("5.14.0-70.58.1.el9_0.x86_64"),
			expectedError: "",
		},
		{
			mapping:       buildLiteralKernelMapping(""),
			expectedError: "'mapping' can not be empty nil",
		},
	}

	for _, testcase := range testCases {
		testBuilder := NewModLoaderContainerBuilder("test")
		testBuilder.WithKernelMapping(testcase.mapping)

		if testcase.expectedError != "" {
			assert.Equal(t, testcase.expectedError, testBuilder.errorMsg)
		} else {
			assert.Equal(t, testBuilder.definition.KernelMappings[0], *testcase.mapping)
		}
	}
}

func TestModuleLoaderContainerWithOptions(t *testing.T) {
	testBuilder := NewModLoaderContainerBuilder("test").WithOptions(
		func(builder *ModuleLoaderContainerBuilder) (*ModuleLoaderContainerBuilder, error) {
			return builder, nil
		})
	assert.Equal(t, "", testBuilder.errorMsg)

	testBuilder = NewModLoaderContainerBuilder("test").WithOptions(
		func(builder *ModuleLoaderContainerBuilder) (*ModuleLoaderContainerBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestModuleLoaderContainerWithVersion(t *testing.T) {
	testCases := []struct {
		version       string
		expectedError string
	}{
		{
			version:       "",
			expectedError: "'version' can not be empty",
		},
		{
			version:       "1.1",
			expectedError: "",
		},
	}

	for _, testcase := range testCases {
		testBuilder := NewModLoaderContainerBuilder("test")
		testBuilder.WithVersion(testcase.version)

		if testcase.expectedError != "" {
			assert.Equal(t, testcase.expectedError, testBuilder.errorMsg)
		} else {
			assert.Equal(t, testBuilder.definition.Version, testcase.version)
		}
	}
}

func TestModuleLoaderContainerBuildModuleLoaderContainerCfg(t *testing.T) {
	testCases := []struct {
		name          string
		expectedError string
		mutate        bool
	}{
		{
			name:          "kmod",
			expectedError: "",
			mutate:        false,
		},
		{
			name:          "",
			expectedError: "'modName' cannot be empty",
			mutate:        false,
		},
		{
			name:          "kmod",
			expectedError: "'mapping' can not be empty nil",
			mutate:        true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewModLoaderContainerBuilder(testCase.name)

		if testCase.mutate {
			testBuilder.WithKernelMapping(nil)
		}

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
		assert.NotNil(t, testBuilder.definition)

		if testCase.expectedError == "" || testCase.name != "" {
			assert.Equal(t, testCase.name, testBuilder.definition.Modprobe.ModuleName)
		}
	}
}

func buildRegExKernelMapping(regexp string) *v1beta1.KernelMapping {
	reg := NewRegExKernelMappingBuilder(regexp)
	regexBuild, _ := reg.BuildKernelMappingConfig()

	return regexBuild
}

func buildLiteralKernelMapping(literal string) *v1beta1.KernelMapping {
	lit := NewLiteralKernelMappingBuilder(literal)
	litBuild, _ := lit.BuildKernelMappingConfig()

	return litBuild
}
