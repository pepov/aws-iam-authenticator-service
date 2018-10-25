package util

import (
	"io/ioutil"
)

// CreateTemporaryFile creates a temporary file with the given content
func CreateTemporaryFile(prefix, content string) (string, string, error) {
	dir, err := ioutil.TempDir("", prefix)
	if err != nil {
		return "", "", err
	}
	configFile := dir + "/config"
	err = ioutil.WriteFile(configFile, []byte(content), 0600)
	if err != nil {
		return "", "", err
	}
	return dir, configFile, nil
}
