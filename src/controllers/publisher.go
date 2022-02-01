package controllers

import (
	"sync"

	"github.com/gusluker/server-bot/src/models"

	log "github.com/sirupsen/logrus"
)

type ControllerObserver interface {
	Run(data *models.Data)
}

type ControllerPublisher struct {
	Subscribers []ControllerObserver
	PuMutex	sync.Mutex
}

func NewControllerPublisher(subscribers []ControllerObserver) *ControllerPublisher {
	retval := &ControllerPublisher {}
	for i := range subscribers {
		retval.Subscribe(subscribers[i])
	}

	return retval
}

func (publisher *ControllerPublisher) Subscribe(subscriber ControllerObserver) {
	publisher.Subscribers = append(publisher.Subscribers, subscriber)
}

func (publisher *ControllerPublisher) Notify(data *models.Data) {
	publisher.PuMutex.Lock()
	log.Debugf("Publisher; LLamando a los suscriptores")
	for i := range publisher.Subscribers {
		publisher.Subscribers[i].Run(data)
	}
	publisher.PuMutex.Unlock()
}

