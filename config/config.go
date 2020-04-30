package config

/*
import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func parseConfigFile() error {
	path, err := getConfigFile()
	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	json.Unmarshal(bytes, &cfg)

	return nil
}

func readConfigFile(path string) ([]byte, error) {
}

// getConfigFile tries to find a JSON configuration file,
// and returns the path to it, if successful.
// It checks in two places:
//  1. in the current directory for config.json file
//  2. in the env var ATLAS_RPM_INSTALLER_CONFIG
// Any values from a parsed config file are overridden by the command line.
func getConfigFile() (string, error) {
	here, err := os.Getwd()
	if err != nil {
		return "", err
	}

	path := filepath.Join(here, "config.json")
	exists, err := fileExists(path)
	if err != nil || exists {
		return path, err
	}

	var found bool
	path, found = os.LookupEnv("ATLAS_RPM_INSTALLER_CONFIG")
	if found {
		return path, nil
	}

	// No config file available
	return "", nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
*/
