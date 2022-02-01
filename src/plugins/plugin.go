package plugins

import (
	"github.com/gusluker/server-bot/src/models"	
)

type Plugin interface {
	IsThisPlugin(data *models.Data)	bool
	Run(data *models.Data)			([]*models.GData, error)
	GetName() 						string
}
