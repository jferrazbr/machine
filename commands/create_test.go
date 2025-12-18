package commands

import (
	"testing"

	"flag"

	"github.com/rancher/machine/commands/commandstest"
	"github.com/rancher/machine/libmachine/mcnflag"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestValidateSwarmDiscoveryErrorsGivenInvalidURL(t *testing.T) {
	err := validateSwarmDiscovery("foo")
	assert.Error(t, err)
}

func TestValidateSwarmDiscoveryAcceptsEmptyString(t *testing.T) {
	err := validateSwarmDiscovery("")
	assert.NoError(t, err)
}

func TestValidateSwarmDiscoveryAcceptsValidFormat(t *testing.T) {
	err := validateSwarmDiscovery("token://deadbeefcafe")
	assert.NoError(t, err)
}

type fakeFlagGetter struct {
	flag.Value
	value interface{}
}

func (ff fakeFlagGetter) Get() interface{} {
	return ff.value
}

var nilStringSlice []string

var getDriverOptsFlags = []mcnflag.Flag{
	mcnflag.BoolFlag{
		Name: "bool",
	},
	mcnflag.IntFlag{
		Name: "int",
	},
	mcnflag.IntFlag{
		Name:  "int_defaulted",
		Value: 42,
	},
	mcnflag.StringFlag{
		Name: "string",
	},
	mcnflag.StringFlag{
		Name:  "string_defaulted",
		Value: "bob",
	},
	mcnflag.StringSliceFlag{
		Name: "stringslice",
	},
	mcnflag.StringSliceFlag{
		Name:  "stringslice_defaulted",
		Value: []string{"joe"},
	},
}

var getDriverOptsTests = []struct {
	data     map[string]interface{}
	expected map[string]interface{}
}{
	{
		expected: map[string]interface{}{
			"bool":                  false,
			"int":                   0,
			"int_defaulted":         42,
			"string":                "",
			"string_defaulted":      "bob",
			"stringslice":           nilStringSlice,
			"stringslice_defaulted": []string{"joe"},
		},
	},
	{
		data: map[string]interface{}{
			"bool":             fakeFlagGetter{value: true},
			"int":              fakeFlagGetter{value: 42},
			"int_defaulted":    fakeFlagGetter{value: 37},
			"string":           fakeFlagGetter{value: "jake"},
			"string_defaulted": fakeFlagGetter{value: "george"},
			// NB: StringSlices are not flag.Getters.
			"stringslice":           []string{"ford"},
			"stringslice_defaulted": []string{"zaphod", "arthur"},
		},
		expected: map[string]interface{}{
			"bool":                  true,
			"int":                   42,
			"int_defaulted":         37,
			"string":                "jake",
			"string_defaulted":      "george",
			"stringslice":           []string{"ford"},
			"stringslice_defaulted": []string{"zaphod", "arthur"},
		},
	},
}

func TestGetDriverOpts(t *testing.T) {
	for _, tt := range getDriverOptsTests {
		commandLine := &commandstest.FakeCommandLine{
			LocalFlags: &commandstest.FakeFlagger{
				Data: tt.data,
			},
		}
		driverOpts := getDriverOpts(commandLine, getDriverOptsFlags)
		assert.Equal(t, tt.expected["bool"], driverOpts.Bool("bool"))
		assert.Equal(t, tt.expected["int"], driverOpts.Int("int"))
		assert.Equal(t, tt.expected["int_defaulted"], driverOpts.Int("int_defaulted"))
		assert.Equal(t, tt.expected["string"], driverOpts.String("string"))
		assert.Equal(t, tt.expected["string_defaulted"], driverOpts.String("string_defaulted"))
		assert.Equal(t, tt.expected["stringslice"], driverOpts.StringSlice("stringslice"))
		assert.Equal(t, tt.expected["stringslice_defaulted"], driverOpts.StringSlice("stringslice_defaulted"))
	}
}

// TestConvertMcnFlagsToCliFlags_TableDriven is a comprehensive table-driven test
// for the convertMcnFlagsToCliFlags function that covers all flag types and edge cases.
func TestConvertMcnFlagsToCliFlags_TableDriven(t *testing.T) {
	tests := []struct {
		name               string
		inputFlags         []mcnflag.Flag
		expectedFlagCount  int
		validationFunction func(t *testing.T, cliFlags []cli.Flag)
		expectedError      bool
	}{
		{
			name: "BoolFlag with Value true should convert to BoolTFlag",
			inputFlags: []mcnflag.Flag{
				&mcnflag.BoolFlag{
					Name:  "enable-feature",
					Usage: "Enable the feature",
					Value: true,
				},
			},
			expectedFlagCount: 1,
			validationFunction: func(t *testing.T, cliFlags []cli.Flag) {
				assert.Equal(t, 1, len(cliFlags), "should have exactly 1 flag")
				boolTFlag, isBoolTFlag := cliFlags[0].(cli.BoolTFlag)
				assert.True(t, isBoolTFlag, "should be converted to BoolTFlag when Value=true")
				assert.Equal(t, "enable-feature", boolTFlag.Name, "flag name should match")
			},
			expectedError: false,
		},
		{
			name: "BoolFlag with Value false should convert to BoolFlag",
			inputFlags: []mcnflag.Flag{
				&mcnflag.BoolFlag{
					Name:  "disable-feature",
					Usage: "Disable the feature",
					Value: false,
				},
			},
			expectedFlagCount: 1,
			validationFunction: func(t *testing.T, cliFlags []cli.Flag) {
				assert.Equal(t, 1, len(cliFlags), "should have exactly 1 flag")
				boolFlag, isBoolFlag := cliFlags[0].(cli.BoolFlag)
				assert.True(t, isBoolFlag, "should be converted to BoolFlag when Value=false")
				assert.Equal(t, "disable-feature", boolFlag.Name, "flag name should match")
			},
			expectedError: false,
		},
		{
			name: "StringFlag should be converted correctly",
			inputFlags: []mcnflag.Flag{
				&mcnflag.StringFlag{
					Name:  "region",
					Usage: "AWS region",
					Value: "us-east-1",
				},
			},
			expectedFlagCount: 1,
			validationFunction: func(t *testing.T, cliFlags []cli.Flag) {
				assert.Equal(t, 1, len(cliFlags), "should have exactly 1 flag")
				_, isStringFlag := cliFlags[0].(cli.StringFlag)
				assert.True(t, isStringFlag, "should be converted to StringFlag")
			},
			expectedError: false,
		},
		{
			name: "IntFlag should be converted correctly",
			inputFlags: []mcnflag.Flag{
				&mcnflag.IntFlag{
					Name:  "port",
					Usage: "Server port",
					Value: 8080,
				},
			},
			expectedFlagCount: 1,
			validationFunction: func(t *testing.T, cliFlags []cli.Flag) {
				assert.Equal(t, 1, len(cliFlags), "should have exactly 1 flag")
				_, isIntFlag := cliFlags[0].(cli.IntFlag)
				assert.True(t, isIntFlag, "should be converted to IntFlag")
			},
			expectedError: false,
		},
		{
			name: "StringSliceFlag should be converted correctly",
			inputFlags: []mcnflag.Flag{
				&mcnflag.StringSliceFlag{
					Name:  "tags",
					Usage: "Resource tags",
					Value: []string{"tag1", "tag2"},
				},
			},
			expectedFlagCount: 1,
			validationFunction: func(t *testing.T, cliFlags []cli.Flag) {
				assert.Equal(t, 1, len(cliFlags), "should have exactly 1 flag")
				_, isStringSliceFlag := cliFlags[0].(cli.StringSliceFlag)
				assert.True(t, isStringSliceFlag, "should be converted to StringSliceFlag")
			},
			expectedError: false,
		},
		{
			name: "Multiple flags of different types should all convert",
			inputFlags: []mcnflag.Flag{
				&mcnflag.StringFlag{
					Name:  "region",
					Usage: "AWS region",
					Value: "us-east-1",
				},
				&mcnflag.BoolFlag{
					Name:  "enable-ssl",
					Usage: "Enable SSL",
					Value: true,
				},
				&mcnflag.IntFlag{
					Name:  "retries",
					Usage: "Number of retries",
					Value: 3,
				},
			},
			expectedFlagCount: 3,
			validationFunction: func(t *testing.T, cliFlags []cli.Flag) {
				assert.Equal(t, 3, len(cliFlags), "should have 3 flags")
				flagNameMap := make(map[string]bool)
				for _, f := range cliFlags {
					// Extract flag name based on type
					switch flag := f.(type) {
					case cli.StringFlag:
						flagNameMap[flag.Name] = true
					case cli.BoolTFlag:
						flagNameMap[flag.Name] = true
					case cli.BoolFlag:
						flagNameMap[flag.Name] = true
					case cli.IntFlag:
						flagNameMap[flag.Name] = true
					}
				}
				assert.True(t, flagNameMap["region"], "region flag should exist")
				assert.True(t, flagNameMap["enable-ssl"], "enable-ssl flag should exist")
				assert.True(t, flagNameMap["retries"], "retries flag should exist")
			},
			expectedError: false,
		},
		{
			name:              "Empty flag list should return empty list",
			inputFlags:        []mcnflag.Flag{},
			expectedFlagCount: 0,
			validationFunction: func(t *testing.T, cliFlags []cli.Flag) {
				assert.Equal(t, 0, len(cliFlags), "should return empty list for empty input")
			},
			expectedError: false,
		},
		{
			name: "BoolFlag with environment variable should preserve EnvVar",
			inputFlags: []mcnflag.Flag{
				&mcnflag.BoolFlag{
					Name:   "feature-enabled",
					Usage:  "Feature enable flag",
					Value:  true,
					EnvVar: "MY_FEATURE_ENABLED",
				},
			},
			expectedFlagCount: 1,
			validationFunction: func(t *testing.T, cliFlags []cli.Flag) {
				assert.Equal(t, 1, len(cliFlags), "should have 1 flag")
				boolTFlag := cliFlags[0].(cli.BoolTFlag)
				assert.Equal(t, "MY_FEATURE_ENABLED", boolTFlag.EnvVar, "EnvVar should be preserved")
			},
			expectedError: false,
		},
		{
			name: "StringFlag with environment variable should preserve EnvVar",
			inputFlags: []mcnflag.Flag{
				&mcnflag.StringFlag{
					Name:   "api-key",
					Usage:  "API key",
					Value:  "default-key",
					EnvVar: "API_KEY",
				},
			},
			expectedFlagCount: 1,
			validationFunction: func(t *testing.T, cliFlags []cli.Flag) {
				assert.Equal(t, 1, len(cliFlags), "should have 1 flag")
				stringFlag := cliFlags[0].(cli.StringFlag)
				assert.Equal(t, "API_KEY", stringFlag.EnvVar, "EnvVar should be preserved")
			},
			expectedError: false,
		},
		{
			name: "Azure managed disks flag should convert correctly",
			inputFlags: []mcnflag.Flag{
				&mcnflag.BoolFlag{
					Name:   "azure-managed-disks",
					Usage:  "Configures VM and availability set for managed disks",
					Value:  true,
					EnvVar: "AZURE_MANAGED_DISKS",
				},
			},
			expectedFlagCount: 1,
			validationFunction: func(t *testing.T, cliFlags []cli.Flag) {
				assert.Equal(t, 1, len(cliFlags), "should have 1 flag")
				_, isBoolTFlag := cliFlags[0].(cli.BoolTFlag)
				assert.True(t, isBoolTFlag, "azure-managed-disks with Value=true should be BoolTFlag")
			},
			expectedError: false,
		},
		{
			name: "BoolFlag without explicit Value should default to false and use BoolFlag",
			inputFlags: []mcnflag.Flag{
				&mcnflag.BoolFlag{
					Name:  "default-bool-flag",
					Usage: "A BoolFlag without explicit Value",
				},
			},
			expectedFlagCount: 1,
			validationFunction: func(t *testing.T, cliFlags []cli.Flag) {
				assert.Equal(t, 1, len(cliFlags), "should have 1 flag")
				_, isBoolFlag := cliFlags[0].(cli.BoolFlag)
				assert.True(t, isBoolFlag, "BoolFlag without explicit Value should be converted to cli.BoolFlag")
			},
			expectedError: false,
		},
		{
			name: "IntFlag without Value should preserve zero value",
			inputFlags: []mcnflag.Flag{
				&mcnflag.IntFlag{
					Name:  "zero-int-flag",
					Usage: "An IntFlag with zero value",
				},
			},
			expectedFlagCount: 1,
			validationFunction: func(t *testing.T, cliFlags []cli.Flag) {
				assert.Equal(t, 1, len(cliFlags), "should have 1 flag")
				intFlag := cliFlags[0].(cli.IntFlag)
				assert.Equal(t, 0, intFlag.Value, "IntFlag without Value should preserve zero value")
			},
			expectedError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(testingContext *testing.T) {
			cliFlags, err := convertMcnFlagsToCliFlags(testCase.inputFlags)

			if testCase.expectedError {
				assert.Error(testingContext, err, "conversion should produce an error")
				return
			}

			assert.NoError(testingContext, err, "conversion should not produce an error")
			assert.Equal(testingContext, testCase.expectedFlagCount, len(cliFlags), "flag count should match expected count")

			if testCase.validationFunction != nil {
				testCase.validationFunction(testingContext, cliFlags)
			}
		})
	}
}
