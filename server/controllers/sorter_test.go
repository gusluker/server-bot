package controllers

import (
	"fmt"
	"time"
	"strings"
	"testing"
	"math/rand"
	"encoding/json"

	"github.com/gusluker/server-bot/server/models"

	log "github.com/sirupsen/logrus"
)

type TestSorterType struct {
	Request string
	Err bool
}

type Observer1 struct {
	V chan string
}

func (ob *Observer1) Run(data *models.Data) {
	if data.NamePlug == "observer1" {
		body, _ := data.Body.(map[string]interface{})
		v, _ := body["value"].(string)
		ob.V <- v
	}
}

var (
	R = rand.New(rand.NewSource(time.Now().UnixNano()))
	testTableSorter = []TestSorterType {
		{
			Request: "{" +
				"\"index\":\"prueba 1\"," +
				"\"plug\":\"observer1\"," +
				"\"coord\": {" +
					"\"lat\": \"6.152279808716283\"," +
					"\"lon\": \"-75.3776536561246\"" +
				"}," +
				"\"body\": {" +
					"\"val\": \"" + fmt.Sprintf("%d", R.Intn(50)) + "\"" +
				"}" +
			"}",	
			Err: false,
		},
		{
			Request: "{" + 
				"\"index\":\"prueba 2\"," +
				"\"plug\":\"observer1\"," +
				"\"coord\": {" + 
					"\"lat\": \"6.206413162163033\"," +
					"\"lon\": \"-75.57156851778012\"" +
				"}," +
				"\"body\": {" +
					"\"val\": \"" + fmt.Sprintf("%d", R.Intn(50)) + "\"" +
				"}" +
			"}",
			Err: false,
		},
		{
			Request: "{" + 
				"\"index\":\"prueba 3\"," +
				"\"plug\":\"observer1\"," +
				"\"body\": {" +
					"\"val\": \"" + fmt.Sprintf("%d", R.Intn(50)) + "\"" +
				"}" +
			"}",
			Err: true,
		},
		{
			Request: "{" + 
				"\"index\":\"prueba 4\"," +
				"\"plug\":\"observer1\"," +
				"\"location\": {" +
					"\"coord\":{ " +
						"\"lat\":\"3.4225175065856273\"," +
						"\"lon\":\"-76.55495718367443\"" +
					"}," +
					"\"country\":\"colombia\"," +
					"\"country_code\":\"co\"," +
					"\"state\":\"valle del cauca\"," +
					"\"city\":\"cali\"" +
				"}," +
				"\"body\": {" +
					"\"val\": \"" + fmt.Sprintf("%d", R.Intn(50)) + "\"" +
				"}" +
			"}",
			Err: false,
		},
	}
)

func TestSorter(t *testing.T) {
	log.SetLevel(log.DebugLevel)	
	ob1 := &Observer1 {
		V: make(chan string, len(testTableSorter)),
	}
	var observers []ControllerObserver
	observers = append(observers, ob1)
	controller := New(observers)

	for i := range testTableSorter {
		if code := sorterTest(testTableSorter[i].Request, controller); testTableSorter[i].Err {
			if code == 200 {
				t.Fatalf("Ciclo %d. Sorter tenía que fallar", i)
			}
		} else if code != 200 {
			t.Fatalf("Ciclo %d. Sorter no tenía que fallar", i)
		}
	}

	for i := range testTableSorter {
		if testTableSorter[i].Err {
			continue
		}

		select {
		case v := <- ob1.V:
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(testTableSorter[i].Request), &data); err != nil {
				t.Fatalf("No se pudo convertir Request")	
			}

			body, _ := data["body"].(map[string]interface{})

			va, _ := body["value"].(string)
			if va != v {
				t.Fatalf("El valor retornado no concuerda. Esperado %s == %s Recibido", va, v)
			}
		case <- time.After(7000 * time.Millisecond):
			t.Fatalf("Ciclo %d. No se recibió datos, faltan %d datos por recibir", i, len(testTableSorter) - i)
		}
	}
}

func sorterTest(request string, controller *Controller) (int){
	sor := &Sor {
		Cont: controller,
		Body: strings.NewReader(request),
		ContentType: "application/json",
	}

	code, _ := sor.sorter()
	return code
}
