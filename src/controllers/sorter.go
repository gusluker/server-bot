package controllers

import (
	"io"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/gusluker/server-bot/src/models"
)

type Sor struct {
	Body 		io.Reader	
	ContentType	string
	Cont		*Controller
}

type Controller struct {
	Publisher	*ControllerPublisher
	Channel 	chan *models.Data
	Quit 		chan int
}

func New(subscribers []ControllerObserver) *Controller {
	controller := &Controller {
		Channel: make(chan *models.Data, 50),
		Quit: make(chan int),
		Publisher: NewControllerPublisher(subscribers),
	}

	go controller.work()
	return controller
}

func (controller *Controller) work() {
	for {
		select {
		case data := <- controller.Channel:
			if data.Loc == nil {
				go controller.process(data)
			} else {
				controller.Publisher.Notify(data)
			}
		case <- controller.Quit:
			return
		}
	}
}

func (controller *Controller) process(data *models.Data) {
	var loc *models.Location
	var err error

	log.Debugf("Sorter; GeoReverse Lat=%s Lon=%s", data.Coord.Latitude, data.Coord.Longitude)
	if loc, err = models.GetLocation(data.Coord); err == nil {
		data.Loc = loc
	}

	if err == nil {
		controller.Publisher.Notify(data)
	} else {
		log.Errorf("No se pudo obtener la geolocalización del dato %s: %s", data.Index, err)
	}
}

func (controller *Controller) Exit() {
	controller.Quit <- 1	
}

func (controller *Controller) Sorter() (func(*gin.Context)) {
	return func(c *gin.Context) {
		sor := &Sor {
			Cont: controller,
			Body: c.Request.Body,
			ContentType: c.GetHeader("Content-Type"),
		}

		code, msg := sor.sorter()
		c.String(code, msg)
	}
}

func (sor *Sor) sorter() (int, string) {
	code := http.StatusBadRequest
	msg := "Content-Type no válido"

	if sor.ContentType == "application/json" {
		if body, err := ioutil.ReadAll(sor.Body); err == nil {
			log.Debugf("Sorter; data: %s", string(body))
			var dataJson map[string]interface{}

			if err := json.Unmarshal(body, &dataJson); err != nil {
				log.Warnf("No se recibió una Request tipo JSON")
				log.Debugf("Sorter; No se recibió una Request tipo JSON; %s", err)
			} else if data, err := models.ToData(dataJson); err == nil{
				switch code = sor.Cont.workQueue(data); code {
					case http.StatusOK:	
						msg = "OK"
					case http.StatusInternalServerError:
						msg = "Internal Server Error"
						log.Warn("No se pudo procesar Request. Cola de trabajo llena")
				}
			} else {
				msg = "internal error"
				code = http.StatusInternalServerError
				log.Errorf("Falló conversión a Data; %s", msg)
				log.Debugf("Sorter; Falló conversión a Data; %s: %s", msg, err)
			}
		} else {
			msg = "bad request"
			log.Error("Error de lectura de Request")
			log.Debugf("Sorter; Error de lectura de Request; %s", err)
		}
	}

	return code, msg
}

func (controller *Controller) workQueue(data *models.Data) int {
	var retval int	

	select {
	case controller.Channel <- data:
		log.Debugf("Sorter; Dato \"%s\" en cola de trabajo", data.Index)
		retval = http.StatusOK
	default:
		retval = http.StatusServiceUnavailable
	}

	return retval
}
