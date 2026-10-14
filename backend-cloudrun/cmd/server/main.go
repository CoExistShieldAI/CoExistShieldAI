package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"ai.coexistshield/backend/internal/api"
	"ai.coexistshield/backend/internal/elephant"
	"ai.coexistshield/backend/internal/gemini"
	"ai.coexistshield/backend/internal/geofence"
	"ai.coexistshield/backend/internal/memory"
	"ai.coexistshield/backend/internal/ml"
	"ai.coexistshield/backend/internal/tiger"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	geminiClient := gemini.NewClient()
	geofenceEngine := geofence.NewEngine()
	memoryBank := memory.NewBank()
	mlPipeline := ml.NewRetrainingPipeline()
	elephantEngine := elephant.NewEngine()
	tigerEngine := tiger.NewEngine()

	router := api.NewRouter(geminiClient, geofenceEngine, memoryBank, mlPipeline, elephantEngine, tigerEngine)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("CoExist Shield AI Cloud Run backend starting on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server listen failed: %v", err)
	}
}
