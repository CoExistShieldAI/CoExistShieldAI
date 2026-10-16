package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"ai.coexistshield/backend/internal/api"
	"ai.coexistshield/backend/internal/auth"
	"ai.coexistshield/backend/internal/compliance"
	"ai.coexistshield/backend/internal/corroboration"
	"ai.coexistshield/backend/internal/elephant"
	"ai.coexistshield/backend/internal/gemini"
	"ai.coexistshield/backend/internal/geofence"
	"ai.coexistshield/backend/internal/memory"
	"ai.coexistshield/backend/internal/ml"
	"ai.coexistshield/backend/internal/repository"
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
	complianceEngine := compliance.NewEngine()
	corroborationEngine := corroboration.NewEngine(geminiClient)

	router := api.NewRouter(geminiClient, geofenceEngine, memoryBank, mlPipeline, elephantEngine, tigerEngine, complianceEngine, corroborationEngine)

	projectID := os.Getenv("AUTH_PROJECT_ID")
	if projectID == "" {
		projectID = os.Getenv("FIREBASE_PROJECT_ID")
	}
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	if projectID == "" {
		projectID = os.Getenv("GCP_PROJECT")
	}
	if projectID == "" {
		// Attempt auto-detection from GCP Metadata Server (Cloud Run / GCE)
		client := &http.Client{Timeout: 800 * time.Millisecond}
		req, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/project/project-id", nil)
		if err == nil {
			req.Header.Set("Metadata-Flavor", "Google")
			if resp, err := client.Do(req); err == nil && resp.StatusCode == http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				projectID = strings.TrimSpace(string(body))
			}
		}
	}

	if projectID != "" {
		router.SetAuthVerifier(auth.NewVerifier(projectID, ""))
		log.Printf("Firebase ID token verification enabled for project %s", projectID)

		firestoreRepo := repository.NewFirestoreRepository(projectID)
		router.SetRepositories(firestoreRepo, firestoreRepo)
		router.InitAdminPersistence(projectID)
		log.Printf("Firestore persistence enabled for project %s (including admin datasets: wildlife, env, weather, roadkill, geofences)", projectID)
	} else if emulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST"); emulatorHost != "" {
		firestoreRepo := repository.NewFirestoreRepository("demo-coexist")
		router.SetRepositories(firestoreRepo, firestoreRepo)
		router.InitAdminPersistence("demo-coexist")
		log.Printf("Firestore persistence enabled via emulator at %s", emulatorHost)
	} else {
		log.Printf("AUTH_PROJECT_ID not set; running in open demo mode with in-memory persistence")
	}

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
