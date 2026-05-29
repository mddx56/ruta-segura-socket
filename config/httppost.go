package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func StringToFloat(input string) float64 {
	floatValue, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0
	}
	return floatValue
}

func StringToInt(str string) int {
	intValue, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}

	return intValue
}

type Position struct {
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Timestamp      int     `json:"timestamp"`
	Speed          float64 `json:"speed"`
	Batt           float64 `json:"batt"`
	Dispositivo_id int     `json:"dispositivo_id"`
}

func NewPosition(latitude, longitude float64, timestamp int, speed, batt float64, dispositivo_id int) Position {
	return Position{
		Latitude:       latitude,
		Longitude:      longitude,
		Timestamp:      timestamp,
		Speed:          speed,
		Batt:           batt,
		Dispositivo_id: dispositivo_id,
	}
}

func DataToPosition(data string) Position {

	parts := strings.Split(data, " ")

	if len(parts) >= 3 {
		// Extraer la URL
		urlStr := parts[1]

		// Analizar la URL
		u, err := url.Parse(urlStr)
		if err != nil {
			fmt.Println("Error al analizar la URL:", err)
			fmt.Println("la URL:", u.String())
		}
		// Obtener el método HTTP y la ruta
		//httpMethod := parts[0]
		//urlPath := u.Path

		// Obtener los parámetros
		queryParams := u.Query()

		// Imprimir el resultado
		//fmt.Printf("Método HTTP: %s\n", httpMethod)
		//fmt.Printf("Ruta URL: %s\n", urlPath)
		// Acceder a los valores individuales
		id := StringToInt(queryParams.Get("id"))
		lat := StringToFloat(queryParams.Get("lat"))
		lon := StringToFloat(queryParams.Get("lon"))
		timestamp := StringToInt(queryParams.Get("timestamp"))
		speed := StringToFloat(queryParams.Get("speed"))
		batt := StringToFloat(queryParams.Get("batt"))

		return Position{
			Latitude:       lat,
			Longitude:      lon,
			Timestamp:      timestamp,
			Speed:          speed,
			Batt:           batt,
			Dispositivo_id: id,
		}
	} else {
		fmt.Println("La cadena de entrada no tiene el formato esperado.")
		return Position{}
	}
}

func SendPosition(data string, posturl string) {
	requestData := DataToPosition(data)
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		fmt.Println("Error cast Json: ", err)
		return
	}
	inCty := radioAllow(requestData.Latitude, requestData.Longitude)
	if inCty {
		fmt.Println("Dentro del Departamento. 🏍")
	} else {
		fmt.Println("Fuera de la Departamento. ☠")
	}
	allow := validatePositionDecimals(requestData.Latitude, requestData.Longitude, 6)

	if allow {

		body := []byte(jsonData)

		fmt.Println("Body:", string(body))

		r, err := http.NewRequest("POST", posturl, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error al enviar posicion: ", err)
			return
		}

		r.Header.Add("Content-Type", "application/json")

		client := &http.Client{}
		res, err := client.Do(r)
		if err != nil {
			fmt.Println("Error peticion http: ", err)
			return
		}

		defer res.Body.Close()

		fmt.Println("Http response:", res.StatusCode)
	} else {
		fmt.Println("Lat, Lng deve ser mayor a 6")
	}
}
