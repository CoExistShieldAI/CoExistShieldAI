package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ai.coexistshield/backend/internal/compliance"
	"ai.coexistshield/backend/internal/corroboration"
	"ai.coexistshield/backend/internal/elephant"
	"ai.coexistshield/backend/internal/gemini"
	"ai.coexistshield/backend/internal/geofence"
	"ai.coexistshield/backend/internal/memory"
	"ai.coexistshield/backend/internal/ml"
	"ai.coexistshield/backend/internal/model"
	"ai.coexistshield/backend/internal/tiger"
)

type Router struct {
	mux                 *http.ServeMux
	geminiClient        *gemini.Client
	geofenceEngine      *geofence.Engine
	memoryBank          *memory.Bank
	mlPipeline          *ml.RetrainingPipeline
	elephantEngine      *elephant.Engine
	tigerEngine         *tiger.Engine
	complianceEngine    *compliance.Engine
	corroborationEngine *corroboration.Engine
}

func NewRouter(
	geminiClient *gemini.Client,
	geofenceEngine *geofence.Engine,
	memoryBank *memory.Bank,
	mlPipeline *ml.RetrainingPipeline,
	elephantEngine *elephant.Engine,
	tigerEngine *tiger.Engine,
	complianceEngine *compliance.Engine,
	corroborationEngine *corroboration.Engine,
) *Router {
	r := &Router{
		mux:                 http.NewServeMux(),
		geminiClient:        geminiClient,
		geofenceEngine:      geofenceEngine,
		memoryBank:          memoryBank,
		mlPipeline:          mlPipeline,
		elephantEngine:      elephantEngine,
		tigerEngine:         tigerEngine,
		complianceEngine:    complianceEngine,
		corroborationEngine: corroborationEngine,
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
	r.mux.HandleFunc("/api/v1/geofences/active", r.handleGetActiveGeofences)
	r.mux.HandleFunc("/api/v1/geofences/recalculate", r.handleGeofencesRecalculate)
	r.mux.HandleFunc("/api/v1/geofences/scent-plume", r.handleGeneratePlumeGeofences)
	r.mux.HandleFunc("/api/v1/meteorology/current", r.handleGetCurrentMeteorology)
	r.mux.HandleFunc("/api/v1/analytics/japan-kaki/top-hazards", r.handleTopHazards)
	r.mux.HandleFunc("/api/v1/compliance/rspo/package", r.handleRspoPackage)
	r.mux.HandleFunc("/api/v1/compliance/rspo/export", r.handleRspoExport)
	r.mux.HandleFunc("/api/v1/compliance/gcf/budget-impact", r.handleGcfBudgetImpact)
	r.mux.HandleFunc("/api/v1/municipal/kaki/priority-layer", r.handlePriorityKakiLayer)
	r.mux.HandleFunc("/api/v1/municipal/kaki/action", r.handleMunicipalKakiAction)
	r.mux.HandleFunc("/api/v1/memory/search", r.handleMemorySearch)
	r.mux.HandleFunc("/api/v1/alerts/explain", r.handleAlertExplanation)
	r.mux.HandleFunc("/api/v1/ml/status", r.handleMlStatus)
	r.mux.HandleFunc("/api/v1/ml/retrain", r.handleMlRetrain)
	r.mux.HandleFunc("/api/v1/ml/predict", r.handleMlPredict)
	r.mux.HandleFunc("/api/v1/ml/query", r.handleMlQuery)
	r.mux.HandleFunc("/api/v1/analytics/thailand-hec/crop-plots", r.handleThailandCropPlots)
	r.mux.HandleFunc("/api/v1/analytics/thailand-hec/recalculate", r.handleThailandHecRecalculate)
	r.mux.HandleFunc("/api/v1/analytics/thailand-hec/plots/", r.handleThailandPlotAction)
	r.mux.HandleFunc("/api/v1/analytics/sumatra-tiger/pastures", r.handleTigerPastures)
	r.mux.HandleFunc("/api/v1/analytics/sumatra-tiger/depredation/report", r.handleTigerDepredationReport)
	r.mux.HandleFunc("/api/v1/analytics/sumatra-tiger/recalculate", r.handleTigerRecalculate)
	r.mux.HandleFunc("/api/v1/analytics/sumatra-tiger/pastures/", r.handleTigerPastureAction)
	r.mux.HandleFunc("/api/v1/nlp/ingest/unstructured", r.handleUnstructuredNlpIngest)
	r.mux.HandleFunc("/api/v1/corroboration/streams", r.handleCorroborationStreams)
	r.mux.HandleFunc("/api/v1/corroboration/ingest", r.handleCorroborationIngest)
	r.mux.HandleFunc("/api/v1/corroboration/evaluate", r.handleCorroborationEvaluate)
	r.mux.HandleFunc("/api/v1/corroboration/poll-cycle", r.handleCorroborationPollCycle)
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

func (r *Router) handleGetActiveGeofences(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	lat, err := strconv.ParseFloat(req.URL.Query().Get("lat"), 64)
	if err != nil {
		lat = 39.7020
	}
	lon, err := strconv.ParseFloat(req.URL.Query().Get("long"), 64)
	if err != nil {
		lon = 141.1520
	}
	radiusKm, err := strconv.ParseFloat(req.URL.Query().Get("radiusKm"), 64)
	if err != nil || radiusKm <= 0 {
		radiusKm = 25.0
	}

	zones := r.geofenceEngine.GetActiveGeofences(lat, lon, radiusKm)
	writeJSON(w, http.StatusOK, zones)
}

func (r *Router) handleGetCurrentMeteorology(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	region := req.URL.Query().Get("region")
	if region == "" {
		region = "JAPAN"
	}

	meteo := r.geofenceEngine.GetCurrentMeteorology(region)
	writeJSON(w, http.StatusOK, meteo)
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

func (r *Router) handleRspoPackage(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	plantationId := req.URL.Query().Get("plantationId")
	if plantationId == "" {
		plantationId = "wilmar_sumatra_02"
	}

	pkg, err := r.complianceEngine.GetRspoAuditPackage(plantationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, pkg)
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

	csvData, err := r.complianceEngine.GetRspoAuditCsv(plantationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"rspo_audit_%s.csv\"", plantationId))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(csvData))
}

func (r *Router) handleGcfBudgetImpact(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	muniCode := req.URL.Query().Get("municipalityCode")
	if muniCode == "" {
		muniCode = "03201"
	}

	resp, err := r.complianceEngine.GetGcfBudgetImpact(muniCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) handlePriorityKakiLayer(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	muniCode := req.URL.Query().Get("municipalityCode")
	if muniCode == "" {
		muniCode = "03201"
	}

	resp, err := r.complianceEngine.GetPriorityKakiLayer(muniCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) handleMunicipalKakiAction(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var actionReq compliance.MunicipalKakiActionRequest
	if err := json.NewDecoder(req.Body).Decode(&actionReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	resp, err := r.complianceEngine.LogMunicipalAction(actionReq)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "forbidden") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, resp)
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

func (r *Router) handleMlStatus(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	active := r.mlPipeline.ActiveModel()
	history := r.mlPipeline.RetrainingHistory()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "HEALTHY",
		"activeModel": active,
		"history":     history,
	})
}

type MlRetrainApiRequest struct {
	Trigger string `json:"trigger"`
	Force   bool   `json:"force"`
}

func (r *Router) handleMlRetrain(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var retrainReq MlRetrainApiRequest
	if req.Body != nil && req.ContentLength > 0 {
		_ = json.NewDecoder(req.Body).Decode(&retrainReq)
	}
	if retrainReq.Trigger == "" {
		retrainReq.Trigger = "admin_manual"
	}

	result, err := r.mlPipeline.Retrain(req.Context(), retrainReq.Trigger, retrainReq.Force)
	if err != nil {
		http.Error(w, fmt.Sprintf("Retraining pipeline error: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

type MlPredictApiRequest struct {
	Latitude               float64 `json:"latitude"`
	Longitude              float64 `json:"longitude"`
	DistNearestPersimmonM  float64 `json:"distNearestPersimmonM"`
	DistNearestCarcassM    float64 `json:"distNearestCarcassM"`
	DistNearestCaneFruitM  float64 `json:"distNearestCaneFruitM"`
	WindScentAlignmentDeg  float64 `json:"windScentAlignmentDeg"`
	MoonDarknessPct        float64 `json:"moonDarknessPct"`
	PrecipitationMm24h      float64 `json:"precipitationMm24h"`
	HistoricalIncidents30d float64 `json:"historicalIncidents30d"`
}

func (r *Router) handleMlPredict(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var predReq MlPredictApiRequest
	if err := json.NewDecoder(req.Body).Decode(&predReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON request: %v", err), http.StatusBadRequest)
		return
	}

	h3Index := ml.GeoToH3Res8(predReq.Latitude, predReq.Longitude)
	fv := ml.FeatureVector{
		H3Index:                h3Index,
		DistNearestPersimmonM:  predReq.DistNearestPersimmonM,
		DistNearestCarcassM:    predReq.DistNearestCarcassM,
		DistNearestCaneFruitM:  predReq.DistNearestCaneFruitM,
		WindScentAlignmentDeg:  predReq.WindScentAlignmentDeg,
		MoonDarknessPct:        predReq.MoonDarknessPct,
		PrecipitationMm24h:      predReq.PrecipitationMm24h,
		HistoricalIncidents30d: predReq.HistoricalIncidents30d,
	}

	activeModel := r.mlPipeline.ActiveModel()
	predResult := activeModel.Predict(fv)

	writeJSON(w, http.StatusOK, predResult)
}

func (r *Router) handleMlQuery(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := ml.DefaultBigQueryConfig()
	sql := ml.GenerateBigQueryFeatureExtractionSQL(cfg)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"query":            sql,
		"resolution":       8,
		"slaTargetSeconds": 45,
	})
}

func (r *Router) handleThailandCropPlots(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	plots := r.elephantEngine.GetRegisteredPlots()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"plots": plots,
		"count": len(plots),
		"monitoredParks": []string{
			"Kui Buri National Park",
			"Kaeng Krachan National Park",
		},
	})
}

type ThailandHecRecalculateRequest struct {
	Weather *model.WeatherConditions `json:"weather,omitempty"`
}

func (r *Router) handleThailandHecRecalculate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var recalcReq ThailandHecRecalculateRequest
	if req.Body != nil && req.ContentLength > 0 {
		_ = json.NewDecoder(req.Body).Decode(&recalcReq)
	}

	weather := model.WeatherConditions{
		WindSpeedKmh:     14.0,
		WindDirectionDeg: 90.0,
		MoonIllumination: 0.15,
	}
	if recalcReq.Weather != nil {
		weather = *recalcReq.Weather
	}

	plots := r.elephantEngine.RecalculateAllPlots(weather)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"plots":          plots,
		"count":          len(plots),
		"recalculatedAt": time.Now().UTC().Format(time.RFC3339),
	})
}

type UpdateHarvestStageRequest struct {
	GrowthStage string                   `json:"growthStage"`
	Weather     *model.WeatherConditions `json:"weather,omitempty"`
}

func (r *Router) handleThailandPlotAction(w http.ResponseWriter, req *http.Request) {
	// Path pattern: /api/v1/analytics/thailand-hec/plots/{id}/harvest-stage
	path := strings.TrimPrefix(req.URL.Path, "/api/v1/analytics/thailand-hec/plots/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "harvest-stage" {
		http.NotFound(w, req)
		return
	}

	plotID := parts[0]
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var updateReq UpdateHarvestStageRequest
	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	updated, err := r.elephantEngine.UpdateHarvestStage(plotID, updateReq.GrowthStage, updateReq.Weather)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (r *Router) handleTigerPastures(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	pastures := r.tigerEngine.GetEnclosures()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"pastures":  pastures,
		"count":     len(pastures),
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
	})
}

type TigerDepredationReportRequestBody struct {
	tiger.ReportDepredationRequest
	Weather *model.WeatherConditions `json:"weather,omitempty"`
}

func (r *Router) handleTigerDepredationReport(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody TigerDepredationReportRequestBody
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	weather := model.WeatherConditions{
		WindSpeedKmh:     10.0,
		WindDirectionDeg: 180.0,
		MoonIllumination: 0.10,
	}
	if reqBody.Weather != nil {
		weather = *reqBody.Weather
	}

	enclosure, alert, err := r.tigerEngine.ReportDepredation(reqBody.ReportDepredationRequest, weather)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enclosure": enclosure,
		"alert":     alert,
	})
}

type TigerRecalculateRequest struct {
	Weather *model.WeatherConditions `json:"weather,omitempty"`
}

func (r *Router) handleTigerRecalculate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var recalcReq TigerRecalculateRequest
	if req.Body != nil {
		_ = json.NewDecoder(req.Body).Decode(&recalcReq)
	}

	weather := model.WeatherConditions{
		WindSpeedKmh:     10.0,
		WindDirectionDeg: 180.0,
		MoonIllumination: 0.10,
	}
	if recalcReq.Weather != nil {
		weather = *recalcReq.Weather
	}

	pastures := r.tigerEngine.RecalculateAll(weather)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"pastures":       pastures,
		"count":          len(pastures),
		"recalculatedAt": time.Now().UTC().Format(time.RFC3339),
	})
}

func (r *Router) handleTigerPastureAction(w http.ResponseWriter, req *http.Request) {
	// Path pattern: /api/v1/analytics/sumatra-tiger/pastures/{id}/clear
	path := strings.TrimPrefix(req.URL.Path, "/api/v1/analytics/sumatra-tiger/pastures/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "clear" {
		http.NotFound(w, req)
		return
	}

	pastureID := parts[0]
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	updated, err := r.tigerEngine.ClearKillSite(pastureID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (r *Router) handleUnstructuredNlpIngest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var nlpReq model.UnstructuredNlpIngestRequest
	if err := json.NewDecoder(req.Body).Decode(&nlpReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON request: %v", err), http.StatusBadRequest)
		return
	}

	channel := nlpReq.SourceChannel
	if channel == "" {
		channel = "UNSTRUCTURED_WEB_FEED"
	}

	entity, err := r.geminiClient.ExtractIncidentEntities(req.Context(), nlpReq.RawText, channel, nlpReq.RegionHint)
	if err != nil {
		http.Error(w, fmt.Sprintf("NLP extraction failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Auto-ingest extracted entity into multi-source corroboration engine
	if r.corroborationEngine != nil && entity.Confidence >= 0.70 {
		region := nlpReq.RegionHint
		if region == "" {
			if entity.Species == "BEAR" {
				region = "JAPAN"
			} else if entity.Species == "ELEPHANT" {
				region = "THAILAND"
			} else if entity.Species == "TIGER" {
				region = "INDONESIA"
			}
		}

		r.corroborationEngine.IngestStreamItem(model.CausalStreamItem{
			ItemId:        fmt.Sprintf("nlp-%s", entity.EntityId),
			StreamType:    model.StreamSocialNLP,
			Region:        region,
			Title:         fmt.Sprintf("NLP: %s %s in %s", entity.Species, entity.IncidentType, entity.LocationDescription),
			Description:   entity.RawText,
			Latitude:      entity.EstimatedLatitude,
			Longitude:     entity.EstimatedLongitude,
			RadiusMeters:  1500,
			Confidence:    entity.Confidence,
			TargetSpecies: entity.Species,
			Timestamp:     entity.ExtractedTimestamp,
			ExpiresAt:     time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
			Metadata: map[string]interface{}{
				"urgency":           entity.UrgencyLevel,
				"attractantContext": entity.AttractantContext,
				"sourceChannel":     entity.SourceChannel,
			},
		})
	}

	resp := model.UnstructuredNlpIngestResponse{
		Entity:    *entity,
		ModelUsed: gemini.GeminiModelVersion,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) handleCorroborationStreams(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	region := req.URL.Query().Get("region")
	streamTypeStr := req.URL.Query().Get("type")

	items := r.corroborationEngine.GetActiveStreamItems(region, model.CausalStreamType(streamTypeStr))
	writeJSON(w, http.StatusOK, items)
}

func (r *Router) handleCorroborationIngest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var item model.CausalStreamItem
	if err := json.NewDecoder(req.Body).Decode(&item); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON request: %v", err), http.StatusBadRequest)
		return
	}

	r.corroborationEngine.IngestStreamItem(item)
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status": "ingested",
		"item":   item,
	})
}

func (r *Router) handleCorroborationEvaluate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var evalReq model.CorroborationRequest
	if err := json.NewDecoder(req.Body).Decode(&evalReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON request: %v", err), http.StatusBadRequest)
		return
	}

	resp, err := r.corroborationEngine.Corroborate(req.Context(), evalReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Corroboration evaluation failed: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) handleCorroborationPollCycle(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	ingestedCount := r.corroborationEngine.RunIngestionCycle()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "cycle_completed",
		"itemsIngested": ingestedCount,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
