package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	if store == nil {
		store = NewStore()
	}
	return &Handler{store: store}
}

// RegisterRoutes registers all admin endpoints on the provided ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Dashboard stats
	mux.HandleFunc("/api/v1/admin/dashboard/stats", h.handleDashboardStats)

	// Geofences
	mux.HandleFunc("/api/v1/admin/geofences", h.handleGeofences)
	mux.HandleFunc("/api/v1/admin/geofences/manual/all", h.handleClearAllManualGeofences)
	mux.HandleFunc("/api/v1/admin/geofences/recalculate", h.handleRecalculateGeofences)
	mux.HandleFunc("/api/v1/admin/geofences/", h.handleGeofenceDetail)

	// Wildlife Incidents
	mux.HandleFunc("/api/v1/admin/incidents/wildlife", h.handleWildlife)
	mux.HandleFunc("/api/v1/admin/incidents/wildlife/", h.handleWildlifeDetail)

	// Roadkills
	mux.HandleFunc("/api/v1/admin/incidents/roadkill", h.handleRoadkill)
	mux.HandleFunc("/api/v1/admin/incidents/roadkill/", h.handleRoadkillDetail)

	// Environmental Data
	mux.HandleFunc("/api/v1/admin/environmental", h.handleEnvironmental)
	mux.HandleFunc("/api/v1/admin/environmental/", h.handleEnvironmentalDetail)

	// Operations & Emergency
	mux.HandleFunc("/api/v1/admin/operations/broadcast", h.handleEmergencyBroadcast)
	mux.HandleFunc("/api/v1/admin/export/", h.handleExport)
}

// --- Helper Functions ---

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// --- Dashboard Handlers ---

func (h *Handler) handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodOptions {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	stats := h.store.GetDashboardStats()
	writeJSON(w, http.StatusOK, stats)
}

// --- Geofence Handlers ---

func (h *Handler) handleGeofences(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		source := r.URL.Query().Get("source")
		geos := h.store.ListGeofences(source)
		writeJSON(w, http.StatusOK, geos)
	case http.MethodPost:
		var g AdminGeofence
		if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}
		g = h.store.CreateGeofence(g)
		writeJSON(w, http.StatusCreated, g)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleGeofenceDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/geofences/")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if h.store.DeleteGeofence(id) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
		} else {
			http.Error(w, "Geofence not found", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleClearAllManualGeofences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	count := h.store.ClearAllManualGeofences()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "success",
		"purged_count":  count,
		"message":       fmt.Sprintf("Purged %d manual geofences", count),
	})
}

func (h *Handler) handleRecalculateGeofences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	count, executedAt := h.store.RecalculateAiGeofences()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":             "recalculated",
		"refreshed_ai_fences": count,
		"executed_at":        executedAt.Format(time.RFC3339),
	})
}

// --- Wildlife Handlers ---

func (h *Handler) handleWildlife(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items := h.store.ListWildlife()
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var item AdminWildlifeIncident
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}
		item = h.store.CreateWildlife(item)
		writeJSON(w, http.StatusCreated, item)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleWildlifeDetail(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/incidents/wildlife/"), "/")
	id := parts[0]
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	if len(parts) > 1 && parts[1] == "verify" {
		if r.Method != http.MethodPatch {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Status string `json:"status"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if payload.Status == "" {
			payload.Status = "VERIFIED"
		}
		updated, ok := h.store.UpdateWildlifeStatus(id, payload.Status)
		if !ok {
			http.Error(w, "Wildlife incident not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, updated)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if h.store.DeleteWildlife(id) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// --- Roadkill Handlers ---

func (h *Handler) handleRoadkill(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		includeExcluded := r.URL.Query().Get("include_excluded") == "true"
		items := h.store.ListRoadkills(includeExcluded)
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var rk AdminRoadkillRecord
		if err := json.NewDecoder(r.Body).Decode(&rk); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}
		rk = h.store.CreateRoadkill(rk)
		writeJSON(w, http.StatusCreated, rk)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleRoadkillDetail(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/incidents/roadkill/"), "/")
	id := parts[0]
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	if len(parts) > 1 {
		action := parts[1]
		if action == "exclude" && r.Method == http.MethodPatch {
			var payload struct {
				IsExcluded bool `json:"is_excluded"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				payload.IsExcluded = true
			}
			updated, ok := h.store.SetRoadkillExclusion(id, payload.IsExcluded)
			if !ok {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			writeJSON(w, http.StatusOK, updated)
			return
		}
		if action == "dispose" && r.Method == http.MethodPatch {
			updated, ok := h.store.SetRoadkillDisposed(id)
			if !ok {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			writeJSON(w, http.StatusOK, updated)
			return
		}
	}

	http.Error(w, "Action Not Supported", http.StatusBadRequest)
}

// --- Environmental Handlers ---

func (h *Handler) handleEnvironmental(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items := h.store.ListEnvironmental()
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var env AdminEnvironmentalData
		if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}
		env = h.store.CreateEnvironmental(env)
		writeJSON(w, http.StatusCreated, env)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleEnvironmentalDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/environmental/")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if h.store.DeleteEnvironmental(id) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// --- Operations Handlers ---

func (h *Handler) handleEmergencyBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var req EmergencyBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Mock dispatching push notifications to active subscribers in target zone
	recipients := 142
	if req.Priority == "URGENT" {
		recipients = 285
	}
	resp := EmergencyBroadcastResponse{
		BroadcastID:        fmt.Sprintf("bcast-%d", time.Now().UnixNano()%1000000),
		EstimatedRecipients: recipients,
		DispatchedAt:       time.Now().UTC(),
		Status:             "DISPATCHED",
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	dataset := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/export/")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "geojson"
	}

	data, contentType, err := h.store.ExportDataset(dataset, format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filename := fmt.Sprintf("coexistshield_%s_%s.%s", dataset, time.Now().Format("20060102150405"), format)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
