package models

import (
	"testing"
)

type TestDataType struct {
	Request map[string]interface{}
	Err bool
}

var (
	testTableData = []TestDataType {
		{
			Request: map[string]interface{} {
				"index": "prueba1",
				"plug": "plug1",
				"location": map[string]interface{} {
					"coord": map[string]interface{} {
						"lat":"6.206413162163033",
						"lon":"-75.57156851778012",
					},
					"country":"colombia",
					"country_code":"co",
					"state":"antioquia",
					"city":"medellín",
				},
				"body":"dato cualquiera",
			},
			Err: false,
		},
		{
			Request: map[string]interface{} {
				"index": "prueba2",
				"plug": "plug1",
				"coord": map[string]interface{} {
					"lat":"6.152279808716283",
					"lon":"-75.3776536561246",
				},
				"body": map[string]interface{} {
					"dato": "dato1",
				},
			},
			Err: false,
		},
		{
			Request: map[string]interface{} {
				"index": "prueba3",
				"plug": "plug1",
				"coord": map[string]interface{} {
					"lat":"3.4225175065856273",
					"lon":"-76.55495718367443",
				},
			},
			Err: true,
		},
		{
			Request: map[string]interface{} {
				"index": "prueba4",
				"plug": "plug1",
				"coord": map[string]interface{} {
					"lat":"3.4225175065856273",
					"lon":"-76.55495718367443",
				},
			},
			Err: true,
		},
		{
			Request: map[string]interface{} {
				"index": "prueba5",
				"plug": "plug1",
				"body": map[string]interface{} {
					"dato": "dato1",
				},
			},
			Err: true,
		},
	}
)

func TestData(t *testing.T) {
	for i := range testTableData {
		if _, err := ToData(testTableData[i].Request); testTableData[i].Err {
			if err == nil {
				t.Fatalf("Ciclo %d. ToData debía fallar", i)
			}
		} else if err != nil {
			t.Fatalf("Ciclo %d. ToData no debía fallar; %s", i, err)
		}
	}
}
