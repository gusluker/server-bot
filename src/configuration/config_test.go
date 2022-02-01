package configuration

import (
	"os"
	"testing"
)

type TestTableConfig struct {
	Path string
	IsError bool
}

var (
	_pathMails = "github.com/gusluker/server-bot/test/emails1.sbot"
	testTableConfig = []TestTableConfig {
		{
			Path: "",
			IsError: true,
		},
		{
			Path: "github.com/gusluker/server-bot/test/config1.sbot",
			IsError: true,
		},
		{
			Path: "github.com/gusluker/server-bot/test/config2.sbot",
			IsError: false,
		},
	}
)

func TestConfigurationNew(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("Panic; %s", err)	
		}
	}()

	for i := range testTableConfig {
		path := os.Getenv("GOPATH") + "/src/"
		paths := &ConfigPaths {
			ConfigFilePath: path + testTableConfig[i].Path,
			ConfigFileEmails: path + _pathMails, 
		}

		_, err := Init(paths)
		if testTableConfig[i].IsError {
			if err == nil {
				t.Fatalf("Debe ser error, indice %d", i)
			}
		} else if err != nil{
			t.Fatalf("No debe ser error, indice: %d; %s", i, err)
		}
	}
}
