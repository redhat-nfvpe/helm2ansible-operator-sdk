package new

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/redhat-nfvpe/helm2ansible-operator-sdk/internal/pathconfig"
)

func verifyFlags() error {
	if len(helmChartRef) == 0 {
		return fmt.Errorf("Please Specify Helm Chart Reference")
	}
	if strings.ContainsAny(helmChartRef, " ") {
		return fmt.Errorf("Helm Chart Cannot Contain Spaces")
	}
	if len(kind) == 0 {
		return fmt.Errorf("Please Specify Operator Kind")
	}
	if strings.ContainsAny(kind, " ") {
		return fmt.Errorf("Kind Cannot Contain Spaces")
	}
	if err := checkKindString(kind); err != nil {
		return err
	}
	if len(apiVersion) == 0 {
		return fmt.Errorf("Please Specify Operator API Version")
	}
	if !apiVersionMatchesConvention(apiVersion) {
		return fmt.Errorf("api version does not match convention")
	}
	if strings.ContainsAny(apiVersion, " ") {
		return fmt.Errorf("API Version Cannot Contain Spaces")
	}
	return nil
}

func parse(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("please specify operator name")
	}
	operatorName = args[0]
	if len(operatorName) == 0 {
		return fmt.Errorf("operator name should not be empty")
	}
	return nil
}

func checkKindString(kind string) error {
	if strings.ContainsAny(kind, "-") {
		return fmt.Errorf("Kind Cannot Contain '-' Character")
	}
	if kind != strings.Title(kind) {
		return fmt.Errorf("Kind Name Must Be Capitalized")
	}
	return nil
}

func verifyOperatorSDKVersion() error {
	var (
		cmdOut []byte
		err    error
	)
	cmdName := "operator-sdk"
	cmdArgs := []string{"version"}

	// if operator-sdk is not installed, or not updated this will throw an error
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		return fmt.Errorf("unexpected error: %v when verifying operator-sdk version; please install operator-sdk or update to latest version", err)
	}
	// make sure operator-sdk is atleast version 0.13.0 or higher
	if err = matchVersion(&cmdOut); err != nil {
		return fmt.Errorf("unexpected error: %v when verifying operator-sdk version; please update to latest version", err)
	}

	return nil
}

func matchVersion(cmdOut *[]byte) error {
	pattern := regexp.MustCompile(`.*version\: \"+v(\d+.\d+.\d+)\".*, commit\: +(.*), go version\: +(.*)`)
	matches := pattern.FindStringSubmatch(string(*cmdOut))

	if l := len(matches); l != 2+2 {
		return fmt.Errorf("expected three matches, received %d instead (version %v)", l, matches)
	}

	version := matches[1]
	if len(version) == 0 {
		return fmt.Errorf("expected operator-sdk version number, got: %v", version)
	}

	outdated, err := outdatedVersion(version)
	if err != nil {
		return err
	}
	if outdated {
		return fmt.Errorf("operator-sdk version is outdated")
	}

	return nil

}

func outdatedVersion(version string) (bool, error) {

	pattern := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
	matches := pattern.FindStringSubmatch(version)

	if l := len(matches); l != 3+1 {
		return true, fmt.Errorf("expected four matches, received %d instead (version %s)", l, version)
	}

	first := matches[1]
	second := matches[2]

	if len(first) == 0 || len(second) == 0 {
		return true, fmt.Errorf("error parsing version number: %v", version)
	}

	firstInt, err := strconv.Atoi(first)
	if err != nil {
		return true, err
	}
	secondInt, err := strconv.Atoi(second)
	if err != nil {
		return true, err
	}

	if firstInt == 0 {
		if secondInt < 8 {
			return true, nil
		}
	}

	return false, nil
}

// GetBasePathConfig ....
func GetBasePathConfig() (*pathconfig.PathConfig, error) {
	// get the current directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting working directory")
	}

	basePath := cwd
	pathConfig := pathconfig.NewConfig(basePath)

	return pathConfig, nil
}

func apiVersionMatchesConvention(version string) bool {
	pattern := regexp.MustCompile(`^([a-zA-Z0-9]*)\.([a-zA-Z0-9]*)\.([a-zA-Z0-9]*)\/([a-zA-Z0-9]*)$`)
	matches := pattern.FindStringSubmatch(version)
	if len(matches) != 4+1 {
		return false
	}
	return true
}
