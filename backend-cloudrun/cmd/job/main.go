package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"ai.coexistshield/backend/internal/corroboration"
	"ai.coexistshield/backend/internal/gemini"
	"ai.coexistshield/backend/internal/ml"
)

func main() {
	task := flag.String("task", "retrain", "Job task to run: 'retrain', 'corroboration_ingest', 'all'")
	trigger := flag.String("trigger", "cron_scheduled", "Retraining trigger cause: 'cron_scheduled', 'event_accumulated_50', 'manual'")
	force := flag.Bool("force", false, "Force model promotion bypassing regression checks")
	flag.Parse()

	log.Printf("[Cloud Run Job] Starting Automated Job (Task: %s, Trigger: %s, Force: %t)...", *task, *trigger, *force)

	startTime := time.Now()

	switch *task {
	case "corroboration_ingest":
		geminiClient := gemini.NewClient()
		engine := corroboration.NewEngine(geminiClient)
		count := engine.RunIngestionCycle()
		log.Printf("[Cloud Run Job COMPLETE] Corroboration Ingestion Cycle Finished. Ingested %d items in %v.",
			count, time.Since(startTime))

	case "retrain":
		runRetrainingJob(*trigger, *force, startTime)

	case "all":
		geminiClient := gemini.NewClient()
		engine := corroboration.NewEngine(geminiClient)
		count := engine.RunIngestionCycle()
		log.Printf("[Cloud Run Job] Ingested %d causal items.", count)
		runRetrainingJob(*trigger, *force, startTime)

	default:
		log.Fatalf("[Cloud Run Job ERROR] Unknown task '%s'", *task)
	}

	os.Exit(0)
}

func runRetrainingJob(trigger string, force bool, startTime time.Time) {
	pipeline := ml.NewRetrainingPipeline()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := pipeline.Retrain(ctx, trigger, force)
	if err != nil {
		log.Fatalf("[Cloud Run Job ERROR] Retraining job failed: %v", err)
	}

	duration := time.Since(startTime)

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))

	log.Printf("[Cloud Run Job COMPLETE] Duration: %v | Promoted: %t | Active Model: %s | Notes: %s",
		duration, result.PromotedToProd, result.ActiveModelVersion, result.GateEvaluationNotes)

	if !result.GateCheckPassed && !force {
		log.Printf("[Cloud Run Job WARNING] Candidate model failed promotion criteria. Preserving current active production model.")
	}
}
