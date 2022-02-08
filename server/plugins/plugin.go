package plugins

import (
	"github.com/gusluker/server-bot/server/models"	
)

type Plugin interface {
	IsThisPlugin(data *models.Data)	bool
	Run(data *models.Data)			([]*models.GData, error)
	GetName() 						string
}
