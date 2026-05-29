package models

// Estructuras para la API de Dispositivos (GET)
type DeviceResponse struct {
	Status  bool         `json:"status"`
	Message string       `json:"message"`
	Data    []DeviceItem `json:"data"`
}

type DeviceItem struct {
	IMEI           string `json:"imei"`
	Model          string `json:"model"`
	SimPhoneNumber string `json:"sim_phone_number"`
}

// Estructura para enviar la Posición (POST)
type PositionPayload struct {
	DeviceID   string  `json:"device_id"`   // IMEI
	DeviceTime string  `json:"device_time"` // ISO8601
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Speed      int     `json:"speed"`
	Course     int     `json:"course"`
	Attributes *string `json:"attributes"` // Puntero para permitir null
}

// MapUpdatePayload: Estructura ligera para WebSocket (React/Leaflet)
type MapUpdatePayload struct {
	IMEI      string  `json:"id"` // "id" es más corto que "imei" para ahorrar ancho de banda
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Speed     int     `json:"spd"` // km/h
	Course    int     `json:"dir"` // Dirección (grados) para rotar el icono en el mapa
	Battery   int     `json:"bat"` // %
	Ignition  bool    `json:"ign"` // Para colorear el icono (Verde/Gris)
	Status    string  `json:"st"`  // "online", "moving", "idle", "alarm"
	Timestamp string  `json:"ts"`  // Hora de la actualización
}
