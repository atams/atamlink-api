package main

import (
	"log"

	"github.com/atam/atamlink/internal/app"
)

func main() {
	// 1. Buat instance aplikasi baru.
	// Fungsi New() akan mengurus semua inisialisasi.
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// 2. Jalankan aplikasi.
	// Metode Run() akan mengurus start server dan graceful shutdown.
	application.Run()
}