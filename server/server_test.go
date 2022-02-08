package server

import (
	"os"	
	"fmt"
	"time"
	"errors"
	"testing"
	"math/rand"	
	"encoding/json"

	"github.com/gusluker/server-bot/server/models"
	"github.com/gusluker/server-bot/server/plugins"

	"github.com/gin-gonic/gin"
	"github.com/appleboy/gofight/v2"
	log "github.com/sirupsen/logrus"
)

type Plug1 struct {
	NCall int	
}

func (plug *Plug1) GetName() string {
	return "plug1"
}

func (plug *Plug1) IsThisPlugin(data *models.Data) bool {
	return data.NamePlug == "plug1"
}

func (plug *Plug1) Run(data *models.Data) ([]*models.GData, error) {
	Call <- true
	plug.NCall += 1
	var retval []*models.GData

	if data != nil {
		bLoc, _ := json.Marshal(data)
		log.Debugf("Plugin1: %s", string(bLoc))

		var body map[string]interface{}
		ok := false
		if body, ok = data.Body.(map[string]interface{}); !ok {
				return nil, errors.New("Plugin1. La opción Body no es del tipo de dato esperado")
		}

		var vI interface{}
		if vI, ok = body["val"]; !ok {
			return nil, errors.New("Plugin1. En la opción Body, no se encontró la opción val")
		}

		val, _ := vI.(string)

		gd1 := &models.GData {
			ContentType: "text/plain",
			ContentTransferEncoding: "quoted-printable",
			Data: fmt.Sprintf("%s: %s\nLocation: %s->%s", data.Index, val, data.Loc.Country, data.Loc.State),
		}

		retval = append(retval, gd1)
	} else {
		return nil, errors.New("Plugin1. Data es nil")
	}

	return retval, nil
}

type Plug2 struct {
	NCall int	
}

func (plug *Plug2) GetName() string {
	return "plug2"
}

func (plug *Plug2) IsThisPlugin(data *models.Data) bool {
	return data.NamePlug == "plug2"
}

func (plug *Plug2) Run(data *models.Data) ([]*models.GData, error) {
	Call <- true
	plug.NCall += 1

	bLoc, _ := json.Marshal(data.Loc)
	log.Debugf("Plugin2: %s", string(bLoc))

	var body map[string]interface{}
	ok := false
	if body, ok = data.Body.(map[string]interface{}); !ok {
		return nil, errors.New("Plugin2. La opción Body no no es del tipo de dato esperado")
	}

	var vI interface{}
	if vI, ok = body["img"]; !ok {
		return nil, errors.New("Plugin2. En la opción Body, no se encontró la opción img1")
	}

	img, _ := vI.(string)

	gd1 := &models.GData {
		ContentType: "text/plain",
		ContentTransferEncoding: "quoted-printable",
		Data: fmt.Sprintf("%s\nLocation: %s->%s", data.Index, data.Loc.Country, data.Loc.State),
	}

	gd2 := &models.GData {
		ContentType: "image/jpeg",
		ContentTransferEncoding: "base64",
		Name: "imagen1.jpeg",
		Data: img,
	}

	var retval []*models.GData
	retval = append(retval, gd1)
	retval = append(retval, gd2)

	return retval, nil
}

const (
	NREQUEST = 50
)

type IGenerator interface {
	GetSize() int	
	GetDataString(index int) string
}

type Generator struct {
	Data []*models.Data
}

func (gen *Generator) GetDataString(index int) string {
	var retval string
	if index < gen.GetSize() {
		strData, _ := json.Marshal(gen.Data[index])
		retval = string(strData)
	}

	return retval
}

func (gen *Generator) GetSize() int {
	return len(gen.Data)	
}

func NewPlug1RandomGenerator() IGenerator {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	retval := &Generator{}

	for i := 0; i < NREQUEST ; i += 1 {
		lon := -1 * (float32(68) + r.Float32() + float32(r.Intn(8)))//-77 - -68
		lat := 1.8 + float32(r.Intn(3)) + r.Float32()//1.8 - 6	
		retval.Data = append(retval.Data , &models.Data{
			Index: "plug1 " + fmt.Sprintf("%d", i),
			NamePlug: "plug1",
			Coord: &models.GeoCoord {
				Latitude: fmt.Sprintf("%v", lat),
				Longitude: fmt.Sprintf("%v", lon),
			},
			Body: map[string]interface{} {
				"val": fmt.Sprintf("%d", r.Intn(1000000)),
			},
		})
	}

	return retval
}

func NewPlug2RandomGenerator() IGenerator {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	retval := &Generator{}

	for i := 0; i < NREQUEST; i += 1 {
		lon := -1 * (float32(76) + r.Float32() + 0.5 + float32(r.Intn(2)))//-76 - -79.8
		lat := -1 * (r.Float32() + r.Float32())//-2 - 0
		retval.Data = append(retval.Data, &models.Data {
			Index: "plug2 " + fmt.Sprintf("%d", i),
			NamePlug: "plug2",
			Coord: &models.GeoCoord{
				Latitude: fmt.Sprintf("%v", lat),
				Longitude: fmt.Sprintf("%v", lon),
			},
			Body: map[string]interface{} {
				"img": "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",
			},
		})
	}

	return retval
}

var (
	Call = make(chan bool)
	Path = "github.com/gusluker/server-bot/test/config4.sbot"
)

func TestServer(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("PANIC: %s", err)
		}
	}()
	
	path := os.Getenv("GOPATH") + "/src/"
	if _, err := os.Stat(path + "github.com/gusluker/server-bot/test/config4.sbot"); errors.Is(err, os.ErrNotExist) {
		t.Fatalf("El archivo github.com/gusluker/server-bot/test/config4.sbot no se encontró. Observe la documentación y defina este archivo con una configuración válida. Necesita dos correos enlazados a co y ec respectivamente. El archivo config4.sbot se encuentra en gitignore porque maneja información personal sensible.")
	}

	var plugs []plugins.Plugin	
	plug1 := Plug1{}
	plug2 := Plug2{}
	plugs = append(plugs, &plug1)
	plugs = append(plugs, &plug2)

	router := gin.Default()
	New(plugs, router)
	log.SetLevel(log.DebugLevel)

	g1 := NewPlug1RandomGenerator()
	g2 := NewPlug2RandomGenerator()
	fight := gofight.New()
	nFail := 0
	for i := 0; i < NREQUEST; i += 1 {
		fight.POST("/sorter").
			SetDebug(false).
			SetHeader(gofight.H{
				"Content-Type":"application/json",
			}).
			SetBody(g1.GetDataString(i)).
			Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
				if res.Code != 200 {
					nFail += 1	
				}
			})

		fight.POST("/sorter").
			SetDebug(false).
			SetHeader(gofight.H{
				"Content-Type":"application/json",
			}).
			SetBody(g2.GetDataString(i)).
			Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
				if res.Code != 200 {
					nFail += 1	
				}
			})
	}

	nTotal := NREQUEST * 2
	for i := 0; i < nTotal; i += 1 {
		select {
		case <-	Call:
			fmt.Printf("Llamada %d\n", i)
		case <- time.After(20 * time.Second):
			t.Fatalf("TIMEOUT. Se realizaron %d llamadas, faltaron o fallaron %d, rechazadas %d", i, nTotal - i, nFail)
		}
	}
}
