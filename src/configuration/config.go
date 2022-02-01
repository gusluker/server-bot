package configuration

import (
	"os"
	"fmt"
	"errors"
	"io/ioutil"
	"encoding/json"
)

type Config struct {
	Paths *ConfigPaths 
	SettingsList map[string]interface{}
}

type ConfigPaths struct {
	ConfigFilePath string
	ConfigFileEmails string
}

func Init(paths *ConfigPaths) (*Config, error) {
	var err error
	var config *Config
	var settings map[string]interface{}
	var emails []interface{}
	
	if settings, err = getContentMaps(paths.ConfigFilePath); err == nil {
		if emails, err = getContentArray(paths.ConfigFileEmails); err == nil {
			settings["emails"] = emails
			config = &Config {
				Paths: paths,
				SettingsList: settings,
			}
		}
	}

	return config, err
}

func getContentMaps(path string) (map[string]interface{}, error) {
	var err error	
	var settings map[string]interface{}
	var bytesFile []byte

	if bytesFile, err = ioutil.ReadFile(path); err == nil {
		if err = json.Unmarshal(bytesFile, &settings); err != nil {
			err = errors.New(fmt.Sprintf("Error Archivo %s: %s", path, err))
		}
	}

	return settings, err
}

func getContentArray(path string) ([]interface{}, error) {
	var err error	
	var settings []interface{}
	var bytesFile []byte

	if bytesFile, err = ioutil.ReadFile(path); err == nil {
		if err = json.Unmarshal(bytesFile, &settings); err != nil {
			err = errors.New(fmt.Sprintf("Error en Archivo %s: %s", path, err))
		}
	}

	return settings, err
}

func (config *Config) LookupOptionString(name string) (string, bool) {
	var retval string
	var ok bool	
	var optionIn interface{}

	if optionIn, ok = config.SettingsList[name]; ok {
		retval, ok = optionIn.(string)
	}

	return retval, ok
}

func (config *Config) LookupOption(name string) (interface{}, bool) {
	retval, ok := config.SettingsList[name]
	return retval, ok
}

func GetDefaultPath() (*ConfigPaths, error) {
	var err error
	var filePath, fileEmails string
	var ok bool

	if filePath, ok = os.LookupEnv("SBOT_CONFIG_FILE_PATH"); !ok {
		err = errors.New("no se definió la variable SBOT_CONFIG_FILE_PATH")
	} else if fileEmails, ok = os.LookupEnv("SBOT_CONFIG_FILE_EMAILS"); !ok {
		err = errors.New("no se definió la variable SBOT_CONFIG_FILE_EMAILS");
	}

	var retval *ConfigPaths
	if ok {
		retval = &ConfigPaths {
			ConfigFilePath: filePath,
			ConfigFileEmails: fileEmails,
		}
	}

	return retval, err
}

