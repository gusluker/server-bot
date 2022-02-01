package models

import (
	"os"
	"testing"
	"github.com/gusluker/server-bot/src/configuration"
)

func TestSmtpInit(t *testing.T) {
	email, ok := os.LookupEnv("SBOT_TEST_SMPT_EMAIL")
	if !ok {
		t.Fatalf("Defina la variable de estado SBOT_TEST_SMPT_EMAIL con el email al que desea que los mensajes se envíen")
	}

	path, ok := os.LookupEnv("GOPATH")
	if !ok {
		t.Fatalf("No se pudo recuperar la variable de entorno GOPATH")
	}


	path = path + "/src/github.com/gusluker/server-bot/test"
	configPaths := &configuration.ConfigPaths {
		ConfigFilePath: path + "/config2.sbot",
		ConfigFileEmails: path + "/emails.sbot",
	}

	config, err := configuration.Init(configPaths)
	if err != nil {
		t.Log("Cree un archivo en la ruta test/configuration/ con el nombre emails.sbot o fallará la inicialización de la configuración. Mirar la sección Configuración de Correos de envío")
		t.Fatalf("Falló la inicialización de la configuración: %s", err)
	}

	var gClients *GmailClients
	if gClients, err = InitSmtpClients(config); err != nil {
		t.Fatalf("Falló la inicialización del cliente Gmail: %s", err)
	}

	var datos []*GData
	datos = append(datos, &GData {
		ContentType: "text/plain",
		ContentTransferEncoding: "7bit",
		Data: "Texto enviado desde prueba Golang",
	})


	if err = gClients.SendInSequence(email, datos); err != nil {
		for _, c := range gClients.Clients {
			t.Log("Client")
			t.Logf("ClientID: %s", c.Config.ClientID)
			t.Logf("ClientSecret: %s", c.Config.ClientSecret)
			t.Logf("AccessToken: %s", c.Token.AccessToken)
			t.Logf("RefreshToken: %s", c.Token.RefreshToken)
			t.Logf("TokenType: %s\n", c.Token.TokenType)
		}
		t.Fatalf("Falló el envío de Email: %s", err)
	}
}


