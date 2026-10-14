package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"ai.coexistshield/backend/internal/ml"
)

func main() {
	trigger := flag.String("trigger", "cron_scheduled", "Retraining trigger cause: 'cron_scheduled', 'event_accumulated_50', 'manual'")
	force := flag.Bool("force", false, "Force model promotion bypassing regression checks")
	flag.Parse()

	log.Printf("[Cloud Run Job] Starting Automated Continuous Retraining Job (Trigger: %s, Force: %t)...", *trigger, *force)

	startTime := time.Now()
	pipeline := ml.NewRetrainingPipeline()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := pipeline.Retrain(ctx, *trigger, *force)
	if err != nil {
		log.Fatalf("[Cloud Run Job ERROR] Retraining job failed: %v", err)
	}

	duration := time.Since(startTime)

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))

	log.Printf("[Cloud Run Job COMPLETE] Duration: %v | Promoted: %t | Active Model: %s | Notes: %s",
		duration, result.PromotedToProd, result.ActiveModelVersion, result.GateEvaluationNotes)

	if !result.GateCheckPassed && !*force {
		log.Printf("[Cloud Run Job WARNING] Candidate model failed promotion criteria. Preserving current active production model.")
	}

	os.Exit(0)
}
