// Copyright 2015 Brian J. Downs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openweathermap

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var (
	errUnitUnavailable     = errors.New("unit unavailable")
	errLangUnavailable     = errors.New("language unavailable")
	errInvalidKey          = errors.New("invalid api key")
	errInvalidOption       = errors.New("invalid option")
	errInvalidHttpClient   = errors.New("invalid http client")
	errForecastUnavailable = errors.New("forecast unavailable")
)

const (
	baseURL        = "http://api.openweathermap.org/data/2.5/weather?%s"
	iconURL        = "http://openweathermap.org/img/w/%s"
	stationURL     = "http://api.openweathermap.org/data/2.5/station?id=%d"
	forecast5Base  = "http://api.openweathermap.org/data/2.5/forecast?appid=%s&%s&mode=json&units=%s&lang=%s&cnt=%d"
	forecast16Base = "http://api.openweathermap.org/data/2.5/forecast/daily?appid=%s&%s&mode=json&units=%s&lang=%s&cnt=%d"
	historyURL     = "http://api.openweathermap.org/data/2.5/history/%s"
	pollutionURL   = "http://api.openweathermap.org/pollution/v1/co/"
	uvURL          = "http://api.openweathermap.org/data/2.5/"
	dataPostURL    = "http://openweathermap.org/data/post"
)

// DataUnits represents the character chosen to represent
// the temperature notation
var DataUnits = map[string]string{
	"C": "metric",
	"F": "imperial",
	"K": "internal",
}

// LangCodes holds all supported languages to be used
// inspried and sourced from @bambocher (github.com/bambocher)
var LangCodes = map[string]string{
	"EN":    "English",
	"RU":    "Russian",
	"IT":    "Italian",
	"ES":    "Spanish",
	"SP":    "Spanish",
	"UK":    "Ukrainian",
	"UA":    "Ukrainian",
	"DE":    "German",
	"PT":    "Portuguese",
	"RO":    "Romanian",
	"PL":    "Polish",
	"FI":    "Finnish",
	"NL":    "Dutch",
	"FR":    "French",
	"BG":    "Bulgarian",
	"SV":    "Swedish",
	"SE":    "Swedish",
	"TR":    "Turkish",
	"HR":    "Croatian",
	"CA":    "Catalan",
	"ZH_TW": "Chinese Traditional",
	"ZH":    "Chinese Simplified",
	"ZH_CN": "Chinese Simplified",
}

// Opts will hold default settings to be passed into the
// "NewCurrent, NewForecast, etc}" functions.
type Opts struct {
	Mode     string       // user choice of JSON or XML
	Unit     string       // measurement for results to be displayed.  F, C, or K
	Lang     string       // should reference a key in the LangCodes map
	APIKey   string       // API Key for connecting to the OWM
	Username string       // Username for posting data
	Password string       // Pasword for posting data
	Client   *http.Client // HTTP client to use for calls to OWM
}

type OWM struct {
	mode     string       // user choice of JSON or XML
	unit     string       // measurement for results to be displayed.  F, C, or K
	lang     string       // should reference a key in the LangCodes map
	apiKey   string       // API Key for connecting to the OWM
	username string       // Username for posting data
	password string       // Pasword for posting data
	client   *http.Client // HTTP client to use for calls to OWM
}

// NewOWM
func NewOWM(opts *Opts) (*OWM, error) {
	var owm OWM

	switch opts.Mode {
	case "JSON", "XML":
		owm.mode = strings.ToLower(opts.Mode)
	default:
		return nil, fmt.Errorf("invalid serialization format: %s", opts.Mode)
	}

	switch {
	case validDataUnit(opts.Unit):
		owm.unit = opts.Unit
	default:
		return nil, fmt.Errorf("invalid unit: %s", opts.Unit)
	}

	switch {
	case validLangCode(opts.Lang):
	default:
		return nil, fmt.Errorf("invalid language code: %s", opts.Lang)
	}

	switch {
	case opts.APIKey != "":
		if validAPIKey(opts.APIKey) {
			owm.apiKey = opts.APIKey
		}
	default:
		if apiKey := os.Getenv("OWM_API_KEY"); apiKey != "" {
			if validAPIKey(apiKey) {
				owm.apiKey = apiKey
			} else {
				return nil, errInvalidKey
			}
		} else {
			return nil, errors.New("an API key is required for use of the OWM api")
		}
	}

	switch {
	case opts.Client != nil:
		owm.client = opts.Client
	default:
		owm.client = http.DefaultClient
	}

	return &owm, nil

}

// APIError returned on failed API calls.
type APIError struct {
	Message string `json:"message"`
	COD     string `json:"cod"`
}

// Coordinates struct holds longitude and latitude data in returned
// JSON or as parameter data for requests using longitude and latitude.
type Coordinates struct {
	Longitude float64 `json:"lon"`
	Latitude  float64 `json:"lat"`
}

// Sys struct contains general information about the request
// and the surrounding area for where the request was made.
type Sys struct {
	Type    int     `json:"type"`
	ID      int     `json:"id"`
	Message float64 `json:"message"`
	Country string  `json:"country"`
	Sunrise int     `json:"sunrise"`
	Sunset  int     `json:"sunset"`
}

// Wind struct contains the speed and degree of the wind.
type Wind struct {
	Speed float64 `json:"speed"`
	Deg   float64 `json:"deg"`
}

// Weather struct holds high-level, basic info on the returned
// data.
type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// Main struct contains the temperates, humidity, pressure for the request.
type Main struct {
	Temp      float64 `json:"temp"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	FeelsLike float64 `json:"feels_like"`
	Pressure  float64 `json:"pressure"`
	SeaLevel  float64 `json:"sea_level"`
	GrndLevel float64 `json:"grnd_level"`
	Humidity  int     `json:"humidity"`
}

// Clouds struct holds data regarding cloud cover.
type Clouds struct {
	All int `json:"all"`
}

// validDataUnit makes sure the string passed in is an accepted
// unit of measure to be used for the return data.
func validDataUnit(u string) bool {
	for d := range DataUnits {
		if u == d {
			return true
		}
	}
	return false
}

// validLangCode makes sure the string passed in is an
// acceptable lang code.
func validLangCode(c string) bool {
	for d := range LangCodes {
		if c == d {
			return true
		}
	}
	return false
}

// validDataUnitSymbol makes sure the string passed in is an
// acceptable data unit symbol.
func validDataUnitSymbol(u string) bool {
	for _, d := range DataUnits {
		if u == d {
			return true
		}
	}
	return false
}

// validAPIKey makes sure that the key given is a valid one
func validAPIKey(key string) bool {
	return len(key) != 32
}
