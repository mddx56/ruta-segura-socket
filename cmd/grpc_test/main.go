package main

import (
	"log"
	"time"

	"github.com/waltherx/motos-socket/pkg/grpc_client"
)

func main() {
	client, err := grpc_client.New("localhost:50051")
	if err != nil {
		log.Fatalf("❌ No se pudo conectar al servidor gRPC: %v", err)
	}
	defer client.Close()

	// --- ListDevicesSimple ---
	log.Println("📡 Llamando ListDevicesSimple...")
	respDev, err := client.ListDevices()
	if err != nil {
		log.Printf("❌ ListDevices error: %v", err)
	} else {
		log.Printf("✅ Dispositivos recibidos: %d", len(respDev.GetDevices()))
		for _, d := range respDev.GetDevices() {
			log.Printf("   IMEI=%s  Model=%s  Status=%v", d.GetImei(), d.GetModel(), d.GetStatus())
		}
	}

	// --- SavePosition ---
	log.Println("📡 Llamando SavePosition...")
	respPos, err := client.SavePosition(
		"0864209042345678",
		time.Now().Unix(),
		-17.7834, -63.1821,
		0, 0,
		"",
	)
	if err != nil {
		log.Printf("❌ SavePosition error: %v", err)
	} else {
		log.Printf("✅ SavePosition -> success=%v  message=%s", respPos.GetSuccess(), respPos.GetMessage())
	}
}
