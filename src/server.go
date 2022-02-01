package server

import (
	"os"	
	"fmt"
	"errors"
	"runtime"
	"strings"
	"log/syslog"
	log "github.com/sirupsen/logrus"

	"github.com/gusluker/server-bot/src/models"
	"github.com/gusluker/server-bot/src/plugins"
	"github.com/gusluker/server-bot/src/maillist"
	"github.com/gusluker/server-bot/src/controllers"
	"github.com/gusluker/server-bot/src/configuration"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Router		*gin.Engine	
	Config		*configuration.Config
	Mails		*maillist.MailList
	Plugins 	[]plugins.Plugin
	GClients	*models.GmailClients
	Controller 	*controllers.Controller
}

func New(plugs []plugins.Plugin, router *gin.Engine) *Server {
	if plugs == nil {
		panic("Error en la inicialización de ServerBot: Plugins no puede ser null")
	}

	if router == nil {
		panic("Error en la inicialización de ServerBot: El motor GIN no puede ser null")
	}

	srv := &Server {
		Plugins: plugs,
		Router: router,
	}

	var configPaths *configuration.ConfigPaths
	var err error
	if configPaths, err = configuration.GetDefaultPath(); err == nil {
		if srv.Config, err = configuration.Init(configPaths); err == nil {
			if err = srv.InitLog(); err == nil {
				if srv.Mails, err = maillist.Init(srv.Config); err == nil {
					if srv.GClients, err = models.InitSmtpClients(srv.Config); err == nil {
						var observers []controllers.ControllerObserver
						observers = append(observers, srv)
						srv.Controller = controllers.New(observers)

						srv.InitRoutes()
					}
				}
			}
		}
	}

	if err != nil {
		panic(err)
	}

	return srv
}

func (srv *Server) InitRoutes() {
	srv.Router.POST("/sorter", srv.Controller.Sorter())
}

func (srv *Server) InitLog() error {
	var err error
	var level log.Level
	var runt string

	if level, err = srv.initLogGetLevel(); err == nil {
		log.SetLevel(level)

		var path string
		if path, err = srv.initLogGetPath(); err == nil {
			var f *os.File
			if f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				log.SetOutput(f)
				log.Infof("La Ruta \"%s\" será utilizada como salida del LOG", path)
			} 
		} else if runt = runtime.GOOS; runt == "linux" {
			if logging, er := syslog.New(toSyslogLevel(level), "server-bot"); er == nil {
				log.SetOutput(logging)
				log.Warnf("Falló la utilización de la ruta pasada a log-path: %s. Se pasó a utilizar SYSLOG", err)
				err = nil
			} else {
				msg := fmt.Sprintf("Falló la utilización de la ruta pasada a log-path: %s. " + 
					"Falló el intento de utilizar syslog: %s", err, er)
				err = errors.New(msg)
			}
		}
	} 

	return err
}

func (srv *Server) initLogGetLevel() (log.Level, error) {
	var err error
	level := log.InfoLevel

	if logLevelI, ok := srv.Config.SettingsList["log-level"]; ok {
		if logLevel, ok := logLevelI.(string); ok {
			switch strings.ToLower(logLevel) {
				case "trace": 
					level = log.TraceLevel
				case "debug": 
					level = log.DebugLevel
				case "info": 
					level = log.InfoLevel	
				case "warn": 
					level = log.WarnLevel	
				case "error":
					level = log.ErrorLevel	
				case "fatal":
					level = log.FatalLevel
				case "panic":
					level = log.PanicLevel	
				default:
					msg := fmt.Sprintf("La opción log-level del archivo de configuración \"%s\" no posee un valor válido", 
						srv.Config.Paths.ConfigFilePath)
					err = errors.New(msg)
			}
		} else {
			msg := fmt.Sprintf("La opción log-level del archivo de configuración \"%s\" no posee un valor válido", 
				srv.Config.Paths.ConfigFilePath)
			err = errors.New(msg)
		}
	} 

	return level, err
}

func (srv *Server) initLogGetPath() (string, error) {
	var path string	
	var err error

	if _, ok := srv.Config.SettingsList["log-path"]; ok {
		if path, ok = srv.Config.LookupOptionString("log-path"); !ok {
			msg := fmt.Sprintf("La opción log-path de archivo de configuración \"%s\" debe ser una cadena", 
				srv.Config.Paths.ConfigFilePath)
			err = errors.New(msg)
		}
	} else {
		msg := fmt.Sprintf("La opción log-path de archivo de configuración \"%s\" no existe", srv.Config.Paths.ConfigFilePath)
		err = errors.New(msg)
	}

	return path, err
}

func toSyslogLevel(level log.Level) syslog.Priority {
	var retval syslog.Priority

	switch level {
		case log.DebugLevel, log.TraceLevel:
			retval = syslog.LOG_DEBUG
		case log.InfoLevel:
			retval = syslog.LOG_INFO
		case log.WarnLevel:
			retval = syslog.LOG_WARNING	
		case log.ErrorLevel:
			retval = syslog.LOG_ERR	
		case log.FatalLevel:
			retval = syslog.LOG_EMERG
		case log.PanicLevel:
			retval = syslog.LOG_CRIT	
	}

	return retval
}

func (srv *Server) Run(data *models.Data) {
	path := data.Loc.GetPath()
	pathA := strings.Split(path, ".")
	if node, ok := srv.Mails.FindPath(pathA); ok {
		var err error
		if mails, ok := node.GetMailsHierarchically(); ok {
			without := true
			for i := range srv.Plugins {
				if srv.Plugins[i].IsThisPlugin(data) {
					without = false	
					log.Infof("Utilizando el Plugin %s", srv.Plugins[i].GetName())

					var gdata []*models.GData
					if gdata, err = srv.Plugins[i].Run(data); err == nil {
						var to string
						for h := range mails {
							to += mails[h] + ";"
						}

						srv.GClients.SendInSequence(to, gdata)
					} else {
						log.Errorf("Plugin %s: %s", srv.Plugins[i].GetName(), err)
					}

					break
				}
			}

			if without {
				err = errors.New(fmt.Sprintf("No existe plugin para procesar el dato %s", data.Index))
			}
		} else {
			err = errors.New(fmt.Sprintf("No se encontraron correos enlazados en la ruta %s del árbol de correos", path))
		}
	} else {
		log.Errorf("No existe la ruta %s en el árbol de correos", path)
	}
}

