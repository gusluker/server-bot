package models

import (
	"fmt"
	"time"
	"errors"
	"strings"
	"context"
	"crypto/rand"
	"encoding/json"
	"encoding/base64"
	log "github.com/sirupsen/logrus"

	"github.com/gusluker/server-bot/server/configuration"

	"golang.org/x/oauth2"	
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/gmail/v1"
)

var (
	bound string	
)

type GData struct {
	ContentType 			string
	ContentTransferEncoding	string
	Name					string
	Data 					string
}

type GmailClients struct {
	Clients []*GClient	
}

type GClient struct {
	Srv 	*gmail.Service	

	Email	string
	Config 	oauth2.Config
	Token	oauth2.Token	
}

func InitSmtpClients(config *configuration.Config) (*GmailClients, error) {
	var err error	
	var cli *GmailClients
	var clients []*GClient
	if clients, err = checkSmtpSyntax(config.SettingsList); err == nil {
		cli = &GmailClients {
			Clients: clients,
		}

		if err = cli.oauth2GmailService(); err != nil {
			cli = nil
		} else {
			bound = boundary()
		}
	}

	return cli, err
}

func checkSmtpSyntax(config map[string]interface{}) ([]*GClient, error) {
	var err error	
	var retval []*GClient
	var emailsA []interface{}

	ok := false
	if emailsI, okI := config["emails"]; okI {
		emailsA, ok = emailsI.([]interface{})
	}
	
	if !ok {
		return nil, errors.New("Error en la estructura Emails")
	}

	name := "Email"
	emailsIni := true
	for _, email := range emailsA {
		emailsIni = false
		name = "Email"
		if emailName, ok := getSmtpProperty(email, name); ok {
			name = "ClientID"
			if clientId, ok := getSmtpProperty(email, name); ok {
				name = "ClientSecret"
				if clienteSecret, ok := getSmtpProperty(email, name); ok {
					name = "AccessToken"	
					if accessToken, ok := getSmtpProperty(email, name); ok {
						name = "RefreshToken"	
						if refreshToken, ok := getSmtpProperty(email, name); ok {
							name = "TokenType"	
							if token, ok := getSmtpProperty(email, name); ok {
								emailsIni = true	
								client := &GClient {
									Email: emailName,
									Config: oauth2.Config {
										ClientID: clientId,
										ClientSecret: clienteSecret,
									},
									Token: oauth2.Token {
										AccessToken: accessToken,
										RefreshToken: refreshToken,
										TokenType: token,
									},
								}

								retval = append(retval, client)
							}
						}
					}
				} 
			} 
		}


		if !emailsIni {
			break	
		}
	}


	if !emailsIni {
		msg := fmt.Sprintf("Error en la propiedad %s del archivo de configuración de Emails", name)
		err = errors.New(msg)
		retval = nil
	}

	return retval, err
}

func getSmtpProperty(emailI interface{}, name string) (string, bool) {
	var retval string
	ok := false

	if email, okI := emailI.(map[string]interface{}); okI {
		if proI, okI := email[name]; okI {
			retval, ok = proI.(string)
		}
	}

	return retval, ok
}

func (client *GmailClients) oauth2GmailService() error {
	var err error
	for _, cli := range client.Clients {
		cli.Config.Endpoint = google.Endpoint	
		cli.Config.RedirectURL = "http://localhost"
		cli.Token.Expiry = time.Now()	

		var tokenSource = cli.Config.TokenSource(context.Background(), &cli.Token)
		if cli.Srv, err = gmail.NewService(context.Background(), option.WithTokenSource(tokenSource)); err != nil {
			break	
		}
	}

	return err
}

func boundary() string {
	diccionario := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"	
	str := make([]byte, 64)
	_, _ = rand.Read(str)
	for k, v := range str {
		str[k] = diccionario[v % byte(len(diccionario))]
	}

	return string(str)
}

func (clients *GmailClients) SendInSequence(to string, data []*GData) error {
	msg := gdataToSmtpRequest(&GData {
		ContentType: "text/plain",
		ContentTransferEncoding: "8bit",
		Data: "Servicio de redireccionamiento de información ServerBot",
	})

	for _, v := range data {
		msg += gdataToSmtpRequest(v)
	}

	msg += "--"

	msgBody := []byte("Content-Type: multipart/mixed; boundary=" + bound + "\n" +
				"MIME-Version: 1.0\n" +
				"to: " + to + "\n" + 
				"subject: ServerBot. No responda a este mensaje\n\n" + 
				"--" + bound + "\n\n" +
				msg)

	var msgGmail gmail.Message	
	msgGmail.Raw = base64.URLEncoding.EncodeToString(msgBody)
	var err error
	for _, c := range clients.Clients {
		if _, err = c.Srv.Users.Messages.Send("me", &msgGmail).Do(); err == nil {
			log.Infof("Email %s. Se envío Email a destinatarios %s", c.Email, to)
			break
		} else {
			log.Debugf("Email %s. Falló el envío de Email: %s", c.Email, err)
			strErr := err.Error()	
			if in := strings.Index(strErr, "Response"); in != -1 {
				strErr = strErr[in:len(strErr)]
				if in = strings.Index(strErr, "{"); in != -1 {
					strErr = strErr[in:len(strErr)]
					var errorData map[string]interface{}
					if json.Unmarshal([]byte(strErr), &errorData) == nil {
						if eI, ok := errorData["error"]; ok {
							e, _ := eI.(string)	
							if eI, ok = errorData["error_description"]; ok {
								ed, _ := eI.(string)
								err = errors.New(fmt.Sprintf("Error al envío de SMTP: %s, %s", e, ed))
							}
						}
					}
				}
			} 
		}
	}

	return err
}

func gdataToSmtpRequest(data *GData) string {
	msg := "\n"
	msg += "Content-Type: " + data.ContentType
	if data.ContentType == "text/plain" {
		msg += "; charset="	+ string('"') + "UTF-8" + string('"') + "\n"
	} else if data.ContentType == "image/jpeg" {
		msg += "; name=" + string('"') + data.Name + string('"') + "\n"
		msg += "Content-Disposition: attachment; filename=" + string('"') + data.Name + string('"') + "\n"
	}

	msg += "MIME-Version: 1.0\n"
	msg += "Content-Transfer-Encoding: " + data.ContentTransferEncoding + "\n\n" 
	msg += data.Data + "\n\n"
	msg += "--" + bound 

	return msg
}
