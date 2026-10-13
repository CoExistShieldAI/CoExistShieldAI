package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ai.coexistshield/backend/internal/gemini"
	"ai.coexistshield/backend/internal/geofence"
	"ai.coexistshield/backend/internal/memory"
	"ai.coexistshield/backend/internal/model"
)

type Router struct {
	mux            *http.ServeMux
	geminiClient   *gemini.Client
	geofenceEngine *geofence.Engine
	memoryBank     *memory.Bank
}

func NewRouter(geminiClient *gemini.Client, geofenceEngine *geofence.Engine, memoryBank *memory.Bank) *Router {
	r := &Router{
		mux:            http.NewServeMux(),
		geminiClient:   geminiClient,
		geofenceEngine: geofenceEngine,
		memoryBank:     memoryBank,
	}
	r.registerRoutes()
	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Enable CORS for Compose Multiplatform Web app
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	r.mux.ServeHTTP(w, req)
}

func (r *Router) registerRoutes() {
	r.mux.HandleFunc("/healthz", r.handleHealth)
	r.mux.HandleFunc("/api/v1/forensics/triage", r.handleForensicsTriage)
	r.mux.HandleFunc("/api/v1/geofences/recalculate", r.handleGeofencesRecalculate)
	r.mux.HandleFunc("/api/v1/geofences/scent-plume", r.handleGeneratePlumeGeofences)
	r.mux.HandleFunc("/api/v1/analytics/japan-kaki/top-hazards", r.handleTopHazards)
	r.mux.HandleFunc("/api/v1/compliance/rspo/export", r.handleRspoExport)
	r.mux.HandleFunc("/api/v1/memory/search", r.handleMemorySearch)
	r.mux.HandleFunc("/api/v1/alerts/explain", r.handleAlertExplanation)
}

func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "coexistshield-backend-cloudrun",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func (r *Router) handleForensicsTriage(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var triageReq model.ForensicTriageRequest
	if err := json.NewDecoder(req.Body).Decode(&triageReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON request: %v", err), http.StatusBadRequest)
		return
	}

	resp, err := r.geminiClient.TriageFieldPhoto(req.Context(), triageReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Gemini inference error: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) handleGeofencesRecalculate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var recalcReq model.GeofenceRecalculateRequest
	if err := json.NewDecoder(req.Body).Decode(&recalcReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON request: %v", err), http.StatusBadRequest)
		return
	}

	resp := r.geofenceEngine.Recalculate(recalcReq)
	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) handleGeneratePlumeGeofences(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var plumeReq model.GeneratePlumeGeofencesRequest
	if err := json.NewDecoder(req.Body).Decode(&plumeReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON request: %v", err), http.StatusBadRequest)
		return
	}

	resp := r.geofenceEngine.GenerateScentPlumeGeofences(plumeReq)
	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) handleTopHazards(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	muniCode := req.URL.Query().Get("municipalityCode")
	if muniCode == "" {
		muniCode = "03201"
	}

	limit := 50
	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	resp := r.geofenceEngine.GetTopHazardPersimmonTrees(muniCode, limit)
	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) handleRspoExport(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	plantationId := req.URL.Query().Get("plantationId")
	if plantationId == "" {
		plantationId = "wilmar_sumatra_02"
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"rspo_audit_%s.csv\"", plantationId))

	csvData := fmt.Sprintf(`audit_id,plantation_id,year,human_wildlife_fatalities,acoustic_deterrent_activations,verified_safe_passages
AUD-2026-001,%s,2026,0,18,342
`, plantationId)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(csvData))
}

func (r *Router) handleMemorySearch(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var searchReq model.MemorySearchRequest
	if err := json.NewDecoder(req.Body).Decode(&searchReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON request: %v", err), http.StatusBadRequest)
		return
	}

	resp := r.memoryBank.SearchAnalogousIncidents(searchReq)
	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) handleAlertExplanation(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var explainReq model.CausalAlertExplanationRequest
	if err := json.NewDecoder(req.Body).Decode(&explainReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON request: %v", err), http.StatusBadRequest)
		return
	}

	effectiveTier := explainReq.Tier
	elevatedFrom := ""
	var analogs []model.IncidentMemoryRecord

	// Retrieve historical analogs if requested
	if explainReq.RetrieveMemory {
		queryText := fmt.Sprintf("Species: %s | Region: %s | Weather: %+v | Attractants: %+v",
			explainReq.TargetSpecies, explainReq.Region, explainReq.Weather, explainReq.NearbyAttractants)
		searchResp := r.memoryBank.SearchAnalogousIncidents(model.MemorySearchRequest{
			QueryText: queryText,
			Region:    explainReq.Region,
			Species:   explainReq.TargetSpecies,
			TopK:      3,
		})

		analogs = searchResp.Matches
		if searchResp.ElevatedTier != "" && searchResp.ElevatedTier != explainReq.Tier {
			elevatedFrom = explainReq.Tier
			effectiveTier = searchResp.ElevatedTier
		}
	}

	explainReq.Tier = effectiveTier
	explanation, err := r.geminiClient.GenerateExplainableAlert(req.Context(), explainReq, analogs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Explanation synthesis error: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, model.CausalAlertExplanationResponse{
		Explanation:       explanation,
		HistoricalAnalogs: analogs,
		ElevatedFromTier:  elevatedFrom,
		EffectiveTier:     effectiveTier,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
