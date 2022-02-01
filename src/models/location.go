package models

import (
	"io"
	"fmt"
	"time"
	"errors"
	"regexp"
	"strings"
	"net/http"
	"encoding/json"
)

var (
	_regCoordLat = regexp.MustCompile(`^([-+]?)([\d]{1,2})((\.)(\d+))?$`)
	_regCoordLong = regexp.MustCompile(`^([-+]?)([\d]{1,3})((\.)(\d+))?$`)
)

type Location struct {
	Coord 		*GeoCoord	`json:coord`
	Country 	string		`json:country`
	CountryCode string		`json:country_code`
	State		string		`json:state`
	City		string		`json:city`
}

type GeoCoord struct {
	Latitude string		`json:lat`
	Longitude string	`json:lon`
}

func (location *Location) GetPath() string {
	path := location.CountryCode
	if len(location.State) > 0 {
		path += "." + location.State
	}

	if len(location.City) > 0 {
		path += "." + location.City
	}

	return path
}

const URL_SERVICE = "https://nominatim.openstreetmap.org/reverse?lat=%s&lon=%s&format=jsonv2"

func GetLocation(coord *GeoCoord) (*Location, error) {
	var lat, lon string = coord.Latitude, coord.Longitude	
	if !_regCoordLat.MatchString(lat) || !_regCoordLong.MatchString(lon) {
		return nil, errors.New("Latitud o Longitud no válidas")
	}

	var retval *Location	
	tr := &http.Transport { IdleConnTimeout: 250 * time.Millisecond }
	client := &http.Client{ Transport: tr }
	res, err := client.Get(fmt.Sprintf(URL_SERVICE, lat, lon))

	if err == nil {
		defer res.Body.Close()
		if body, errR := io.ReadAll(res.Body); errR == nil {
			var loca map[string]interface{}
			bodyLower := strings.ToLower(string(body))

			if errR = json.Unmarshal([]byte(bodyLower), &loca); errR == nil {
				if retval, err = checkLocationSyntax(loca); err == nil {
					retval.Coord = coord
				}
			} else {
				err = errR
			}
		} else {
			err = errR	
		}
	}

	return retval, err
}

func checkLocationSyntax(data map[string]interface{}) (*Location, error) {
	var retval *Location	
	var err error
	var ok bool
	var da interface{}

	if da, ok = data["address"]; ok {
		if address, ok := da.(map[string]interface{}); ok {
			retval, err = getLocation(address)
		} else {
			err = errors.New("no se recibió respuesta esperada por el servicio de georeverse")
		}
	} else {
		err = errors.New("no se recibió respuesta esperada por el servicio de georeverse")
	}

	return retval, err
}

func getLocation(address map[string]interface{}) (*Location, error) {
	var err error
	var retval *Location

	var city string	
	if city, err = getCity(address); err == nil {
		var state string
		if state, err = getLocationProperty(address, "state"); err == nil {
			var country string
			if country, err = getLocationProperty(address, "country"); err == nil {
				var countryCo string
				if countryCo, err = getLocationProperty(address, "country_code"); err == nil {
					retval = &Location {
						City: strings.ToLower(city),
						State: strings.ToLower(state),
						Country: strings.ToLower(country),
						CountryCode: strings.ToLower(countryCo),
					}
				}
			} 
		}
	}

	return retval, err
}

func getCity(address map[string]interface{}) (string, error) {
	var city string
	var cityI interface{}
	var ok bool

	var err error

	if cityI, ok = address["county"]; ok {
		city, ok = cityI.(string)
	} else if cityI, ok = address["city"]; ok {
		city, ok = cityI.(string)
	} else if cityI, ok = address["town"]; ok {
		city, ok = cityI.(string)
	} else if cityI, ok = address["suburb"]; ok {
		city, ok = cityI.(string)
	} else if cityI, ok = address["village"]; ok {
		city, ok = cityI.(string)
	} else {
		err = errors.New("no existe un parámetro válido para City en Location")
	}

	return city, err 
}

func getCoord(address map[string]interface{}) (*GeoCoord, error) {
	var err error
	var coord *GeoCoord

	if cI, ok := address["coord"]; ok {
		if c, ok := cI.(map[string]interface{}); ok {
			var lat, lon string	
			if lat, err = getLocationProperty(c, "lat"); err == nil {
				if lon, err = getLocationProperty(c, "lon"); err == nil {
					coord = &GeoCoord {
						Latitude: lat,
						Longitude: lon,
					}
				}
			}
		} else {
			err = errors.New("el parámetro Coord de Location no es válido")
		}
	} else {
		err = errors.New("no se encontró el parámetro Coord en Location")
	}

	return coord, err
}

func getLocationProperty(address map[string]interface{}, name string) (string, error) {
	var property string
	var propertyI interface{}
	var ok bool
	var err error

	if propertyI, ok = address[name]; ok {
		if property, ok = propertyI.(string); !ok {
			err = errors.New(fmt.Sprintf("el parámetro %s de Location no es válido", name))
		}
	} else {
		err = errors.New(fmt.Sprintf("no se encontró el parámetro %s en Location", name))
	}

	return property, err
}
