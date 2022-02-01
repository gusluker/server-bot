package models

import (
	"testing"
)

type TestGetLocationType struct {
	coord 	*GeoCoord
	res		*Location
	err 	bool
}

var (
	testTable = []TestGetLocationType{
		{
			coord: &GeoCoord{
				Latitude: "6.206413162163033",
				Longitude: "-75.57156851778012",
			},
			res: &Location{
				Country: "colombia",
				CountryCode: "co",
				State: "antioquia",
				City: "medellín",
			},
			err: false,
		},
		{
			coord: &GeoCoord{
				Latitude: "6.152279808716283",
				Longitude: "-75.3776536561246",
			},
			res: &Location{
				Country: "colombia",
				CountryCode: "co",
				State: "antioquia",
				City: "rionegro",
			},
			err: false,
		},
		{
			coord: &GeoCoord{
				Latitude: "3.4225175065856273",
				Longitude: "-76.55495718367443",
			},
			res: &Location{
				Country: "colombia",
				CountryCode: "co",
				State: "valle del cauca",
				City: "cali",
			},
			err: false,
		},
		{
			coord: &GeoCoord{
				Latitude: "7.955665816413806",
				Longitude: "-73.94824118183332",
			},
			res: &Location{
				Country: "colombia",
				CountryCode: "co",
				State: "bolívar",
				City: "simití",
			},
			err: false,
		},
		{
			coord: &GeoCoord{
				Latitude: "7.955665816413806.77",
				Longitude: "texto",
			},
			res: nil,
			err: true,
		},
		{
			coord: &GeoCoord{
				Latitude: "180.955665816413806.77",
				Longitude: "1000",
			},
			res: nil,
			err: true,
		},
	}
)

func TestGetLocation(t *testing.T) {
	for i := range testTable {
		coord := &GeoCoord {
			Latitude: testTable[i].coord.Latitude,
			Longitude: testTable[i].coord.Longitude,
		}

		loc, err := GetLocation(coord)
		if testTable[i].err {
			if err == nil {
				t.Fatalf("lat=%s,log=%s no retornó error", coord.Latitude, coord.Longitude)
			}
		} else if err != nil {
			t.Fatalf("lat=%s,log=%s retornó un error; %s", coord.Latitude, coord.Longitude, err)
		} else {
			res := testTable[i].res
			formatErr := "lat=%s,log=%s no retornó el %s esperado %s==%s"
			if res.Country != loc.Country {
				t.Fatalf(formatErr, coord.Latitude, coord.Longitude, "Country", res.Country, loc.Country)
			}

			if res.CountryCode != loc.CountryCode {
				t.Fatalf(formatErr, coord.Latitude, coord.Longitude, "CountryCode", res.CountryCode, loc.CountryCode)
			}

			if res.State != loc.State {
				t.Fatalf(formatErr, coord.Latitude, coord.Longitude, "State", res.CountryCode, loc.CountryCode)
			}

			if res.City != loc.City {
				t.Fatalf(formatErr, coord.Latitude, coord.Longitude, "City", res.City, loc.City)
			}
		}
	}
}
