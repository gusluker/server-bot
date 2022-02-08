package models

import (
	"errors"	
)

type Data struct {
	Index		string 		`json:"index"`
	NamePlug	string 		`json:"plug"`
	Coord		*GeoCoord	`json:"coord"`
	Loc			*Location	`json:"location"`
	Body		interface{}	`json:"body"`
}
	
func ToData(data map[string]interface{}) (*Data, error) {
	var retval *Data
	var err error

	var coord *GeoCoord
	var loc *Location
	if _, ok := data["coord"]; ok {
		coord, err = getCoord(data)
	} else if lI, ok := data["location"]; ok {
		if l, ok := lI.(map[string]interface{}); ok {
			if loc, err = getLocation(l); err == nil {
				loc.Coord, err = getCoord(l)
			}
		} else {
			err = errors.New("el parámetro Location no es válido")
		}
	} else {
		err = errors.New("No se encontró el parámetro Coord o Location en la petición")
	}
	
	if err == nil {
		var index, namePlug string
		if index, err = getLocationProperty(data, "index"); err == nil {
			if namePlug, err = getLocationProperty(data, "plug"); err == nil {
				var body interface{}
				if body, err = getBody(data); err == nil {
					retval = &Data {
						Index: index,
						NamePlug: namePlug,
						Loc: loc,
						Coord: coord,
						Body: body,
					}
				}
			}
		}
	}

	return retval, err
}

func getBody(data map[string]interface{}) (interface{}, error) {
	var retval interface{}
	var err error

	ok := false
	if retval, ok = data["body"]; !ok {
		err = errors.New("En data se esperada el parámetro Body")
	}

	return retval, err
}
