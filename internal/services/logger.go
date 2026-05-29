package services

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type LoggerService struct {
	LogURL string
	APIKey string
	client *http.Client
}

func NewLoggerService(logURL, apiKey string) *LoggerService {
	return &LoggerService{
		LogURL: logURL,
		APIKey: apiKey,
		client: &http.Client{Timeout: 3 * time.Second}, // Timeout de 3s para evitar bloqueos
	}
}

func (ls *LoggerService) SendLog(message string) {
	payload := map[string]string{
		"payload": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", ls.LogURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", ls.APIKey)

	resp, err := ls.client.Do(req)
	if err != nil {
		log.Println("Error sending log to API:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("API returned status: %d\n", resp.StatusCode)
	}
}
