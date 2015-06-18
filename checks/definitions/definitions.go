// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package definitions defines the structures to use pre-made checks and custom
// checks in pre-commit-go.yml.
//
// Each of the struct in this file is to be embedded into pre-commit-go.yml.
// Use the comments here as a guidance to set the relevant values.
//
// The config has two root keys, 'version' and 'modes'. The valid values for
// 'modes' are 'pre-commit', 'pre-push', 'continuous-integration' and 'lint'.
// Each mode has two values; checks and max_duration. 'checks' is a list of
// check defined in this mode, 'max_duration' is the maximum duration allowed
// to run all the checks. If runtime exceeds max_duration, the run is marked as
// failed because it is too slow.
//
// Here's a sample pre-commit-go.yml file:
//
//    min_version: 0.4.6
//    modes:
//      continuous-integration:
//        checks:
//          build:
//          - extra_args: []
//          coverage:
//          - use_global_inference: false
//            use_coveralls: false
//            global:
//              min_coverage: 50
//              max_coverage: 100
//            per_dir_default:
//              min_coverage: 0
//              max_coverage: 0
//            per_dir: {}
//          custom:
//          - display_name: sample-pre-commit-go-custom-check
//            description: runs the check sample-pre-commit-go-custom-check on this repository
//            command:
//            - sample-pre-commit-go-custom-check
//            - check
//            check_exit_code: true
//            prerequisites:
//            - help_command:
//              - sample-pre-commit-go-custom-check
//              - -help
//              expected_exit_code: 2
//              url: "github.com/maruel/pre-commit-go/samples/sample-pre-commit-go-custom-check"
//          gofmt:
//          - {}
//          goimports:
//          - {}
//          test:
//          - extra_args:
//            - -v
//            - -race
//        max_duration: 120
//      lint:
//        checks:
//          errcheck:
//          - ignores: Close
//          golint:
//          - blacklist: []
//          govet:
//          - blacklist:
//            - ' composite literal uses unkeyed fields'
//        max_duration: 15
//      pre-commit:
//        checks:
//          build:
//          - extra_args: []
//          gofmt:
//          - {}
//          test:
//          - extra_args:
//            - -short
//        max_duration: 5
//      pre-push:
//        checks:
//          coverage:
//          - use_global_inference: false
//            use_coveralls: false
//            global:
//              min_coverage: 50
//              max_coverage: 100
//            per_dir_default:
//              min_coverage: 0
//              max_coverage: 0
//            per_dir: {}
//          goimports:
//          - {}
//          test:
//          - extra_args:
//            - -v
//            - -race
//        max_duration: 15
//    ignore_patterns:
//    - ".*"
//    - "_*"
//    - "*.pb.go"
//
// To generate the default `pre-commit-go.yml` file, use:
//
//    pre-commit-go writeconfig
//
package definitions

import (
	"os"

	"github.com/maruel/pre-commit-go/internal"
)

// CheckPrerequisite describe a Go package that is needed to run a Check.
//
// It must list a command that is to be executed and the expected exit code to
// verify that the custom tool is properly installed. If the executable is not
// detected, "go get $URL" will be executed.
type CheckPrerequisite struct {
	// HelpCommand is the help command to run to detect if this prerequisite is
	// installed or not. This command should have no adverse effect and must be
	// fast to execute.
	HelpCommand []string `yaml:"help_command"`
	// ExpectedExitCode is the exit code expected when HelpCommand is executed.
	ExpectedExitCode int `yaml:"expected_exit_code"`
	// URL is the url to fetch as `go get URL`.
	URL string `yaml:"url"`
}

// IsPresent returns true if the prerequisite is present on the system.
func (c *CheckPrerequisite) IsPresent() bool {
	_, exitCode, _ := internal.Capture(cwd, nil, c.HelpCommand...)
	return exitCode == c.ExpectedExitCode
}

// Native checks.

// Build builds everything inside the current directory via
// 'go build ./...'.
//
// This check is mostly useful for executables, that is, "package main".
// Packages containing tests are covered via check Test.
//
// Use multiple Build instances to build multiple times with different tags.
type Build struct {
	// ExtraArgs can be used to build with different tags, e.g. to
	// build -tags foo,zoo.
	ExtraArgs []string `yaml:"extra_args"`
}

// Gofmt runs gofmt in check mode with code simplification enabled.
//
// It is almost redundant with goimports except for '-s' which goimports
// doesn't implement and gofmt doesn't require any external package.
//
// Gofmt has no configuration option. -s is always used.
type Gofmt struct {
}

// Test runs all tests via go test.
//
// Use the specialized check Coverage when -cover is desired.
//
// Use multiple Test instances to test multiple times with different flags,
// like with different tags, with or without the race detector, etc.
type Test struct {
	// ExtraArgs can be used to run the test with additional arguments like -v,
	// -short, -race, etc.
	ExtraArgs []string `yaml:"extra_args"`
}

// Non-native checks; running these require installing third party packages.

// Errcheck runs errcheck on all directories containing .go files.
//
// https://github.com/kisielk/errcheck
type Errcheck struct {
	// Ignores is the flag to pass to -ignore.
	Ignores string `yaml:"ignores"`
}

// Goimports runs goimports in check mode.
//
// Goimports has no configuration option.
//
// https://golang.org/x/tools/cmd/goimports
type Goimports struct {
}

// Golint runs golint.
//
// golint triggers false positives by design so it is preferable to use it only
// in lint mode.
//
// https://github.com/golang/lint
type Golint struct {
	// Blacklist causes this check to ignore the messages generated by golint
	// that contain one of the string listed here.
	Blacklist []string `yaml:"blacklist"`
}

// Govet runs "go tool vet".
//
// govet triggers false positives by design so it is preferable to use it only
// in lint mode.
//
// https://golang.org/cmd/vet
type Govet struct {
	// Blacklist causes this check to ignore the messages generated by go tool vet
	// that contain one of the string listed here.
	Blacklist []string `yaml:"blacklist"`
}

// CoverageSettings permits specifying different coverage values than the
// default values for a specific directory.
type CoverageSettings struct {
	// MinCoverage is the minimum test coverage to be generated or the check is
	// considered to fail. The value is in percent.
	MinCoverage float64 `yaml:"min_coverage"`
	// MaxCoverage is the maximum test coverage to be generated or the check is
	// considered to fail. This is meant to create a 'band' to detect when
	// coverage increased enough so the values are updated. It is fine to use
	// 100. and be done with it. The value is in percent. If 0, the value is not
	// enforced.
	MaxCoverage float64 `yaml:"max_coverage"`
}

// Coverage runs all tests with coverage.
//
// Each testable package is run with 'go test -cover' then all coverage
// information is merged together. This means that package X/Y may create code
// coverage for package X/Z.
//
// When running on https://travis-ci.org, it tries to upload code coverage
// results to https://coveralls.io.
//
// Otherwise, only a summary is printed in case code coverage is not above
// t.MinCoverage.
type Coverage struct {
	// UseGlobalInference determines if coverage from any unit test should be
	// considered for coverage calculation for all package. If false, only the
	// package's unit test is used for coverage calculation.
	//
	// When set to false, directories with no tests will not be accounted for.
	UseGlobalInference bool `yaml:"use_global_inference"`
	// UseCoveralls determines if the data should be sent to https://coveralls.io.
	UseCoveralls bool `yaml:"use_coveralls"`
	// Global coverage parameters. The whole coverage must fit these values. This
	// gives a broad range that the code must maintain.
	Global CoverageSettings `yaml:"global"`
	// PerDirDefault is the default coverage settings per package. It can be
	// overriden by PerDir. This gives a tight range each package should adhere
	// to.
	PerDirDefault CoverageSettings `yaml:"per_dir_default"`
	// PerDir defines a list of directories to use different coverage value.  The
	// directories must be against the root repository. The paths must be in
	// POSIX format, e.g. with '/' as directory element separator. The root path
	// is ".".
	//
	// You can disable coverage for a specific directory by specifying coverage:0.
	PerDir map[string]*CoverageSettings `yaml:"per_dir"`
}

// Extensibility.

// Custom represents a user configured check running an external program.
//
// It can be used multiple times to run multiple external checks.
type Custom struct {
	// DisplayName is check's display name, required.
	DisplayName string `yaml:"display_name"`
	// Description is check's description, optional.
	Description string `yaml:"description"`
	// Command is check's command line, required.
	Command []string `yaml:"command"`
	// CheckExitCode specifies if the check is declared to fail when exit code is
	// non-zero.
	CheckExitCode bool `yaml:"check_exit_code"`
	// Prerequisites are check's prerequisite packages to install first before
	// running the check, optional.
	Prerequisites []CheckPrerequisite `yaml:"prerequisites"`
}

// Private stuff.

var cwd string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cwd = wd
}
