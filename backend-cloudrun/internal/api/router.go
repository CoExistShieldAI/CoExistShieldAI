package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"ai.coexistshield/backend/internal/admin"

	"ai.coexistshield/backend/internal/auth"
	"ai.coexistshield/backend/internal/coldstart"
	"ai.coexistshield/backend/internal/compliance"
	"ai.coexistshield/backend/internal/corroboration"
	"ai.coexistshield/backend/internal/elephant"
	"ai.coexistshield/backend/internal/gemini"
	"ai.coexistshield/backend/internal/geofence"
	"ai.coexistshield/backend/internal/memory"
	"ai.coexistshield/backend/internal/ml"
	"ai.coexistshield/backend/internal/model"
	"ai.coexistshield/backend/internal/repository"
	"ai.coexistshield/backend/internal/tiger"
)

type Router struct {
	mux                 *http.ServeMux
	geminiClient        *gemini.Client
	geofenceEngine      *geofence.Engine
	memoryBank          *memory.Bank
	mlPipeline          *ml.RetrainingPipeline
	bigqueryReplicator  *ml.BigQueryGISReplicator
	elephantEngine      *elephant.Engine
	tigerEngine         *tiger.Engine
	complianceEngine    *compliance.Engine
	corroborationEngine *corroboration.Engine
	coldstartEngine     *coldstart.Engine
	verifier            *auth.Verifier
	incidentRepo        repository.IncidentRepository
	roadkillRepo        repository.RoadkillRepository
	adminHandler        *admin.Handler
	mu                  sync.RWMutex
	incidents           []model.IncidentDto
	roadkills           []model.RoadkillRecordDto
	userRoles           map[string]string
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
	memRepo := repository.NewMemoryRepository()
	r := &Router{
		mux:                 http.NewServeMux(),
		geminiClient:        geminiClient,
		geofenceEngine:      geofenceEngine,
		memoryBank:          memoryBank,
		mlPipeline:          mlPipeline,
		bigqueryReplicator:  ml.NewBigQueryGISReplicator(mlPipeline),
		elephantEngine:      elephantEngine,
		tigerEngine:         tigerEngine,
		complianceEngine:    complianceEngine,
		corroborationEngine: corroborationEngine,
		coldstartEngine:     coldstart.NewEngine(mlPipeline, geofenceEngine, memoryBank),
		incidentRepo:        memRepo,
		roadkillRepo:        memRepo,
		adminHandler:        admin.NewHandler(nil),
		userRoles:           make(map[string]string),
	}

	// Dynamic model promotion listener (Issue #21): updates live geofencing engine without server restart
	if mlPipeline != nil && geofenceEngine != nil {
		mlPipeline.RegisterPromotionListener(func(promoted ml.IncursionModel) {
			geofenceEngine.ApplyPromotedWeights(promoted.Version, promoted.Weights)
		})
	}

	// Real-time synchronization: Wire admin console geofence changes to EventHub
	if r.adminHandler != nil && r.geofenceEngine != nil && r.geofenceEngine.GetEventHub() != nil {
		r.adminHandler.SetGeofenceChangeListener(func(action string, geofenceID string, loc *model.ActiveGeofencePoint, tier string) {
			eventType := "UPDATED"
			switch action {
			case "created":
				eventType = "CREATED"
			case "deleted":
				eventType = "DELETED"
			case "recalculated":
				eventType = "RECALCULATED"
			}
			r.geofenceEngine.GetEventHub().Broadcast(model.GeofenceEvent{
				Type:       eventType,
				Action:     action,
				GeofenceID: geofenceID,
				Location:   loc,
				Tier:       tier,
				Message:    fmt.Sprintf("Admin geofence mutation: %s (id: %s)", action, geofenceID),
				Timestamp:  time.Now().UTC().Format(time.RFC3339),
			})
		})
	}

	// Real-time synchronization: Wire admin console wildlife & roadkill incidents to BigQuery GIS
	if r.adminHandler != nil {
		r.adminHandler.SetIncidentListener(func(category string, id string, species string, lat float64, lng float64, status string) {
			if r.bigqueryReplicator != nil && lat != 0 && lng != 0 {
				incursionVal := 0
				if strings.EqualFold(status, "VERIFIED") || category == "WILDLIFE" {
					incursionVal = 1
				}
				streamEvent := ml.IncidentStreamingEvent{
					EventID:           fmt.Sprintf("evt_%s", id),
					IncidentID:        id,
					Species:           species,
					Category:          category,
					Latitude:          lat,
					Longitude:         lng,
					Timestamp:         time.Now().UTC(),
					IsIncursion:       incursionVal,
					VerificationState: status,
				}
				go func() {
					count, err := r.bigqueryReplicator.StreamSyncBatch([]ml.IncidentStreamingEvent{streamEvent})
					if err != nil {
						log.Printf("[BigQuery GIS] Admin stream sync error for %s: %v", id, err)
					} else {
						log.Printf("[BigQuery GIS] Successfully streamed admin incident %s to BigQuery GIS (count: %d)", id, count)
					}
				}()
			}
		})

		// Real-time synchronization: Wire emergency broadcasts to Geofence EventHub (SSE push to all clients)
		r.adminHandler.SetBroadcastListener(func(req admin.EmergencyBroadcastRequest, broadcastID string, recipients int) {
			tier := "TIER_2_WARNING"
			if strings.EqualFold(req.Priority, "URGENT") {
				tier = "TIER_3_DANGER"
			}
			r.geofenceEngine.GetEventHub().Broadcast(model.GeofenceEvent{
				Type:       "EMERGENCY_BROADCAST",
				Action:     "broadcast",
				GeofenceID: req.GeofenceID,
				Tier:       tier,
				Message:    fmt.Sprintf("[%s] %s: %s", req.Priority, req.Title, req.Body),
				Timestamp:  time.Now().UTC().Format(time.RFC3339),
			})
			log.Printf("[Emergency Broadcast] Broadcast %s dispatched (tier: %s, estimated recipients: %d)", broadcastID, tier, recipients)
		})
	}

	// Real-time dynamic GIS attractants provider: binds live Kaki, Carcasses, and Crops to BigQuery GIS
	if r.bigqueryReplicator != nil {
			r.bigqueryReplicator.SetAttractantProvider(func() (kaki []ml.AttractantFeaturePoint, carcasses []ml.AttractantFeaturePoint, crops []ml.AttractantFeaturePoint) {
				// 1. Dynamic Kaki persimmon trees
				if r.complianceEngine != nil {
					if resp, err := r.complianceEngine.GetPriorityKakiLayer("03201"); err == nil && resp != nil {
						for _, t := range resp.PriorityTrees {
							if t.ManagementStatus != "FELLED" {
								kaki = append(kaki, ml.AttractantFeaturePoint{
									Type:      "KAKI",
									Latitude:  t.Latitude,
									Longitude: t.Longitude,
									Weight:    t.HazardScore,
								})
							}
						}
					}
				}

				// 2. Dynamic Carcasses from roadkill records
				r.mu.RLock()
				for _, rk := range r.roadkills {
					if !strings.EqualFold(rk.ClearanceStatus, "CLEARED") && rk.Location.Latitude != 0 {
						carcasses = append(carcasses, ml.AttractantFeaturePoint{
							Type:      "CARCASS",
							Latitude:  rk.Location.Latitude,
							Longitude: rk.Location.Longitude,
							Weight:    rk.AttractantThreatScore,
						})
					}
				}
				r.mu.RUnlock()

				// 3. Dynamic Elephant crops & Tiger kill-sites
				if r.elephantEngine != nil {
					for _, p := range r.elephantEngine.GetRegisteredPlots() {
						if p.GrowthStage != "HARVESTED_CLEARED" {
							crops = append(crops, ml.AttractantFeaturePoint{
								Type:      "CROP",
								Latitude:  p.Latitude,
								Longitude: p.Longitude,
								Weight:    p.ActiveRiskScore,
							})
						}
					}
				}
				if r.tigerEngine != nil {
					for _, enc := range r.tigerEngine.GetEnclosures() {
						if enc.IsActiveKillSite {
							carcasses = append(carcasses, ml.AttractantFeaturePoint{
								Type:      "CARCASS",
								Latitude:  enc.Latitude,
								Longitude: enc.Longitude,
								Weight:    enc.ActiveThreatScore,
							})
						}
					}
				}
				return kaki, carcasses, crops
			})
		}


	// Seed data is ONLY loaded in development / demo mode or test runs (never in bare production)
	if os.Getenv("ENABLE_SEED_DATA") == "true" || os.Getenv("DEV_MODE") == "true" || os.Getenv("ENV") == "test" || os.Getenv("CI") != "" || os.Getenv("AUTH_PROJECT_ID") == "" {
		r.initSeedData()
	}
	r.registerRoutes()
	return r
}

func (r *Router) SetRepositories(incRepo repository.IncidentRepository, rkRepo repository.RoadkillRepository) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if incRepo != nil {
		r.incidentRepo = incRepo
	}
	if rkRepo != nil {
		r.roadkillRepo = rkRepo
	}
}

// InitAdminPersistence initializes Firestore persistence for the admin dashboard datasets
func (r *Router) InitAdminPersistence(projectID string) {
	if r.adminHandler != nil && r.adminHandler.GetStore() != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := r.adminHandler.GetStore().InitPersistence(ctx, projectID); err != nil {
			fmt.Printf("[Router] Warning: failed to initialize admin persistence with Firestore: %v\n", err)
		} else {
			fmt.Printf("[Router] Admin persistence successfully synchronized with Firestore for project %s\n", projectID)
		}
	}
}

// InitDomainPersistence wires Elephant, Tiger, and Compliance domain engines to Firestore
func (r *Router) InitDomainPersistence(projectID string) {
	firestoreRepo := repository.NewFirestoreRepository(projectID)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Elephant Crop Plots
	if r.elephantEngine != nil {
		if rawPlots, err := firestoreRepo.ListCropPlots(ctx); err == nil && len(rawPlots) > 0 {
			plots := make([]elephant.ElephantCropPlot, 0, len(rawPlots))
			for _, m := range rawPlots {
				p := elephant.ElephantCropPlot{
					ID:                    fmt.Sprint(m["id"]),
					Name:                  fmt.Sprint(m["name"]),
					Province:              fmt.Sprint(m["province"]),
					ParkBoundary:          fmt.Sprint(m["parkBoundary"]),
					CropType:              fmt.Sprint(m["cropType"]),
					GrowthStage:           fmt.Sprint(m["growthStage"]),
					Latitude:              toFloat(m["latitude"]),
					Longitude:             toFloat(m["longitude"]),
					AreaHectares:          toFloat(m["areaHectares"]),
					DistanceToForestEdgeM: toFloat(m["distanceToForestEdgeM"]),
					AromaVolatilityIndex:  toFloat(m["aromaVolatilityIndex"]),
					MaturityScore:         toFloat(m["maturityScore"]),
					ActiveRiskScore:       toFloat(m["activeRiskScore"]),
					ScentBufferMeters:     toFloat(m["scentBufferMeters"]),
					UpdatedAt:             fmt.Sprint(m["updatedAt"]),
				}
				plots = append(plots, p)
			}
			r.elephantEngine.LoadPlots(plots)
			log.Printf("[DomainPersistence] Loaded %d elephant crop plots from Firestore", len(plots))
		}
		r.elephantEngine.SetPersistenceListener(func(plot elephant.ElephantCropPlot) {
			pCtx, pCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer pCancel()
			if err := firestoreRepo.SaveCropPlot(pCtx, plot.ID, plot.Name, plot.Province, plot.ParkBoundary, plot.CropType, plot.GrowthStage, plot.UpdatedAt, plot.Latitude, plot.Longitude, plot.AreaHectares, plot.DistanceToForestEdgeM, plot.AromaVolatilityIndex, plot.MaturityScore, plot.ActiveRiskScore, plot.ScentBufferMeters); err != nil {
				log.Printf("[DomainPersistence] Failed to persist crop plot %s: %v", plot.ID, err)
			}
		})
	}

	// 2. Tiger Livestock Enclosures
	if r.tigerEngine != nil {
		if rawEncs, err := firestoreRepo.ListLivestockEnclosures(ctx); err == nil && len(rawEncs) > 0 {
			encs := make([]tiger.LivestockEnclosure, 0, len(rawEncs))
			for _, m := range rawEncs {
				enc := tiger.LivestockEnclosure{
					ID:                    fmt.Sprint(m["id"]),
					Name:                  fmt.Sprint(m["name"]),
					Province:              fmt.Sprint(m["province"]),
					Village:               fmt.Sprint(m["village"]),
					LivestockType:         fmt.Sprint(m["livestockType"]),
					HeadCount:             toInt(m["headCount"]),
					HasNightCorral:        toBool(m["hasNightCorral"]),
					IsActiveKillSite:      toBool(m["isActiveKillSite"]),
					Latitude:              toFloat(m["latitude"]),
					Longitude:             toFloat(m["longitude"]),
					DistanceToForestEdgeM: toFloat(m["distanceToForestEdgeM"]),
					HoursSinceKill:        toFloat(m["hoursSinceKill"]),
					ReturnProbability:     toFloat(m["returnProbability"]),
					ActiveThreatScore:     toFloat(m["activeThreatScore"]),
					ScentBufferMeters:     toFloat(m["scentBufferMeters"]),
					ForensicNotes:         fmt.Sprint(m["forensicNotes"]),
					UpdatedAt:             fmt.Sprint(m["updatedAt"]),
				}
				encs = append(encs, enc)
			}
			r.tigerEngine.LoadEnclosures(encs)
			log.Printf("[DomainPersistence] Loaded %d tiger livestock enclosures from Firestore", len(encs))
		}
		r.tigerEngine.SetPersistenceListener(func(enc tiger.LivestockEnclosure) {
			pCtx, pCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer pCancel()
			if err := firestoreRepo.SaveLivestockEnclosure(pCtx, enc.ID, enc.Name, enc.Province, enc.Village, enc.LivestockType, enc.ForensicNotes, enc.UpdatedAt, enc.HeadCount, enc.HasNightCorral, enc.IsActiveKillSite, enc.Latitude, enc.Longitude, enc.DistanceToForestEdgeM, enc.HoursSinceKill, enc.ReturnProbability, enc.ActiveThreatScore, enc.ScentBufferMeters); err != nil {
				log.Printf("[DomainPersistence] Failed to persist livestock enclosure %s: %v", enc.ID, err)
			}
		})
	}

	// 3. Municipal Kaki Trees & Felling Logs
	if r.complianceEngine != nil {
		if rawTrees, err := firestoreRepo.ListMunicipalKakiTrees(ctx); err == nil && len(rawTrees) > 0 {
			trees := make([]compliance.MunicipalPriorityKakiTree, 0, len(rawTrees))
			for _, m := range rawTrees {
				tree := compliance.MunicipalPriorityKakiTree{
					TreeID:                         fmt.Sprint(m["treeId"]),
					Address:                        fmt.Sprint(m["address"]),
					Latitude:                       toFloat(m["latitude"]),
					Longitude:                      toFloat(m["longitude"]),
					HazardScore:                    toFloat(m["hazardScore"]),
					IsTop5Percent:                  toBool(m["isTop5Percent"]),
					DistanceToSchoolRouteMeters:    toInt(m["distanceToSchoolRouteMeters"]),
					EstimatedBearAttractantRadiusM: toInt(m["estimatedBearAttractantRadiusM"]),
					OwnerConsentStatus:             fmt.Sprint(m["ownerConsentStatus"]),
					PermitID:                       fmt.Sprint(m["permitId"]),
					ManagementStatus:               fmt.Sprint(m["managementStatus"]),
					RecommendedClearanceOrder:      toInt(m["recommendedClearanceOrder"]),
					AssignedSquad:                  fmt.Sprint(m["assignedSquad"]),
					FruitLoad:                      fmt.Sprint(m["fruitLoad"]),
				}
				trees = append(trees, tree)
			}
			r.complianceEngine.LoadTrees(trees)
			log.Printf("[DomainPersistence] Loaded %d municipal kaki trees from Firestore", len(trees))
		}
		r.complianceEngine.SetPersistenceListener(func(tree compliance.MunicipalPriorityKakiTree, logEntry *compliance.FellingAuditLog) {
			pCtx, pCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer pCancel()
			if err := firestoreRepo.SaveMunicipalKakiTree(pCtx, tree.TreeID, tree.Address, tree.OwnerConsentStatus, tree.PermitID, tree.ManagementStatus, tree.AssignedSquad, tree.FruitLoad, tree.Latitude, tree.Longitude, tree.HazardScore, tree.IsTop5Percent, tree.DistanceToSchoolRouteMeters, tree.EstimatedBearAttractantRadiusM, tree.RecommendedClearanceOrder); err != nil {
				log.Printf("[DomainPersistence] Failed to persist kaki tree %s: %v", tree.TreeID, err)
			}
			if logEntry != nil {
				if err := firestoreRepo.SaveFellingAuditLog(pCtx, logEntry.LogID, logEntry.TreeID, logEntry.PermitID, logEntry.OfficerID, logEntry.Action, logEntry.SchoolDistanceM, logEntry.RiskDropPct, logEntry.BiomassKg); err != nil {
					log.Printf("[DomainPersistence] Failed to persist felling audit log %s: %v", logEntry.LogID, err)
				}
			}
		})
	}
	log.Printf("[DomainPersistence] Firestore listeners wired for elephant, tiger, and compliance engines (project: %s)", projectID)
}

// InitMLPersistence hydrates the continuous retraining ML pipeline from Firestore / GCS and registers persistence hooks.
func (r *Router) InitMLPersistence(projectID string) {
	if r.mlPipeline == nil {
		return
	}
	firestoreRepo := repository.NewFirestoreRepository(projectID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Restore latest promoted model checkpoint
	if modelBytes, err := firestoreRepo.GetLatestMLModel(ctx); err == nil && len(modelBytes) > 0 {
		if err := r.mlPipeline.LoadModelFromJSON(modelBytes); err == nil {
			active := r.mlPipeline.ActiveModel()
			log.Printf("[MLPersistence] Loaded active model checkpoint %s (AUC: %.3f) from Firestore", active.Version, active.Metrics.AucRoc)
		} else {
			log.Printf("[MLPersistence] Warning: Failed to parse Firestore model JSON: %v", err)
		}
	} else {
		log.Printf("[MLPersistence] No promoted model in Firestore yet; using baseline model (%s)", r.mlPipeline.ActiveModel().Version)
	}

	// 2. Restore historical retraining results
	if rawHistory, err := firestoreRepo.ListMLRetrainHistory(ctx); err == nil && len(rawHistory) > 0 {
		history := make([]ml.RetrainResult, 0, len(rawHistory))
		for _, raw := range rawHistory {
			var res ml.RetrainResult
			if err := json.Unmarshal([]byte(raw), &res); err == nil {
				history = append(history, res)
			}
		}
		if len(history) > 0 {
			r.mlPipeline.LoadHistory(history)
			log.Printf("[MLPersistence] Loaded %d retraining history records from Firestore", len(history))
		}
	}

	// 3. Restore persisted training samples
	if rawSamples, err := firestoreRepo.ListMLTrainingSamples(ctx); err == nil && len(rawSamples) > 0 {
		samples := make([]ml.TrainingSample, 0, len(rawSamples))
		for _, raw := range rawSamples {
			var s ml.TrainingSample
			if err := json.Unmarshal([]byte(raw), &s); err == nil {
				samples = append(samples, s)
			}
		}
		if len(samples) > 0 {
			total := r.mlPipeline.LoadSamples(samples)
			log.Printf("[MLPersistence] Restored %d training samples from Firestore (total in store: %d)", len(samples), total)
		}
	}

	// 4. Register listener for new training samples
	r.mlPipeline.SetSampleListener(func(sample ml.TrainingSample) {
		pCtx, pCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer pCancel()
		sampleBytes, err := json.Marshal(sample)
		if err != nil {
			return
		}
		sampleID := fmt.Sprintf("sample-%d", time.Now().UnixNano())
		if err := firestoreRepo.SaveMLTrainingSample(pCtx, sampleBytes, sampleID); err != nil {
			log.Printf("[MLPersistence] Failed to persist training sample: %v", err)
		}
	})

	// 5. Register listener for promoted models and retraining results
	modelsBucket := os.Getenv("ML_MODELS_BUCKET")
	if modelsBucket == "" {
		modelsBucket = os.Getenv("GCS_MEDIA_BUCKET")
	}

	r.mlPipeline.SetPersistenceListener(func(active ml.IncursionModel, result ml.RetrainResult) {
		pCtx, pCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer pCancel()

		// Serialize result and save to history
		if resBytes, err := json.Marshal(result); err == nil {
			if err := firestoreRepo.SaveMLRetrainResult(pCtx, resBytes, result.RunID, result.ActiveModelVersion, result.Trigger, result.PromotedToProd, result.TrainedAt.Format(time.RFC3339)); err != nil {
				log.Printf("[MLPersistence] Failed to persist retrain result %s: %v", result.RunID, err)
			}
		}

		// If candidate was promoted, archive model checkpoint
		if result.PromotedToProd {
			modelBytes, err := json.MarshalIndent(result.CandidateModel, "", "  ")
			if err != nil {
				log.Printf("[MLPersistence] Failed to marshal promoted model %s: %v", result.CandidateModel.Version, err)
				return
			}

			// Archive to Firestore
			if err := firestoreRepo.SaveMLModelCheckpoint(pCtx, modelBytes, result.CandidateModel.Version, result.CandidateModel.TrainedAt.Format(time.RFC3339), result.CandidateModel.Metrics.AucRoc, result.CandidateModel.Metrics.PrecisionTop20, result.CandidateModel.TrainingSampleSize); err != nil {
				log.Printf("[MLPersistence] Failed to save promoted model checkpoint to Firestore: %v", err)
			} else {
				log.Printf("[MLPersistence] Successfully archived promoted model checkpoint %s to Firestore", result.CandidateModel.Version)
			}

			// Archive to Google Cloud Storage if bucket configured
			if modelsBucket != "" {
				versionObj := fmt.Sprintf("models/incursion_%s.json", result.CandidateModel.Version)
				latestObj := "models/incursion_latest.json"
				if u, err := admin.UploadToGCS(modelsBucket, versionObj, "application/json", modelBytes); err == nil {
					log.Printf("[MLPersistence] Uploaded model checkpoint to GCS: %s", u)
				} else {
					log.Printf("[MLPersistence] Failed to upload model to GCS (%s): %v", versionObj, err)
				}
				_, _ = admin.UploadToGCS(modelsBucket, latestObj, "application/json", modelBytes)
			}
		}
	})
	log.Printf("[MLPersistence] Firestore & GCS persistence listeners wired for continuous retraining pipeline (project: %s)", projectID)
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0.0
	}
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	default:
		return 0
	}
}

func toBool(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func computeRoadkillThreatScore(species string, decay string) float64 {
	decayBonus := 0.25
	switch strings.ToUpper(decay) {
	case "BLOATED":
		decayBonus = 0.45 // Peak olfactory plume
	case "DECOMPOSING":
		decayBonus = 0.35
	case "FRESH":
		decayBonus = 0.25
	case "SKELETAL":
		decayBonus = 0.10
	}

	speciesLure := 0.20
	upperSpecies := strings.ToUpper(species)
	if strings.Contains(upperSpecies, "BOAR") || strings.Contains(upperSpecies, "DEER") {
		speciesLure = 0.45 // High trophic value lure for apex carnivores
	} else if strings.Contains(upperSpecies, "BEAR") || strings.Contains(upperSpecies, "TIGER") {
		speciesLure = 0.35
	}

	total := decayBonus + speciesLure
	if total > 0.98 {
		total = 0.98
	}
	if total < 0.15 {
		total = 0.15
	}
	return math.Round(total*100.0) / 100.0
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Enable CORS for Web Admin Portal & Compose Multiplatform
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	auth.Middleware(r.verifier, r.mux).ServeHTTP(w, req)
}

// SetAuthVerifier enables Firebase ID token verification. When unset, all routes stay open
// (local/demo mode). Privileged routes then require an OFFICIAL or ADMIN role claim.
func (r *Router) SetAuthVerifier(v *auth.Verifier) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.verifier = v
}

func (r *Router) officialOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		r.mu.RLock()
		v := r.verifier
		r.mu.RUnlock()
		auth.RequireRole(v, h, "OFFICIAL", "ADMIN")(w, req)
	}
}

func (r *Router) adminOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		r.mu.RLock()
		v := r.verifier
		r.mu.RUnlock()
		auth.RequireRole(v, h, "ADMIN")(w, req)
	}
}

func (r *Router) registerRoutes() {
	r.mux.HandleFunc("/healthz", r.handleHealth)
	r.mux.HandleFunc("/health", r.handleHealth)
	r.mux.HandleFunc("/api/v1/health", r.handleHealth)
	r.mux.HandleFunc("/", r.handleRoot)
	r.mux.HandleFunc("/api/v1/forensics/triage", r.handleForensicsTriage)
	r.mux.HandleFunc("/api/v1/incidents", r.handleIncidents)
	r.mux.HandleFunc("/api/v1/roadkills", r.handleGetRoadkills)
	r.mux.HandleFunc("/api/v1/roadkills/", r.officialOnly(r.handleRoadkillAction))
	r.mux.HandleFunc("/api/v1/geofences/active", r.handleGetActiveGeofences)
	r.mux.HandleFunc("/api/v1/geofences/events", r.handleGeofenceEvents)

	r.mux.HandleFunc("/api/v1/geofences/recalculate", r.handleGeofencesRecalculate)
	r.mux.HandleFunc("/api/v1/geofences/scent-plume", r.handleGeneratePlumeGeofences)
	r.mux.HandleFunc("/api/v1/geofences/forecast-morph", r.handleGenerateForecastPlumeGeofences)
	r.mux.HandleFunc("/api/v1/geofences/deescalate", r.handleDeescalateGeofences)
	r.mux.HandleFunc("/api/v1/meteorology/current", r.handleGetCurrentMeteorology)
	r.mux.HandleFunc("/api/v1/analytics/japan-kaki/top-hazards", r.handleTopHazards)
	r.mux.HandleFunc("/api/v1/compliance/rspo/package", r.officialOnly(r.handleRspoPackage))
	r.mux.HandleFunc("/api/v1/compliance/rspo/export", r.officialOnly(r.handleRspoExport))
	r.mux.HandleFunc("/api/v1/compliance/gcf/budget-impact", r.handleGcfBudgetImpact)
	r.mux.HandleFunc("/api/v1/municipal/kaki/priority-layer", r.handlePriorityKakiLayer)
	r.mux.HandleFunc("/api/v1/municipal/kaki/action", r.officialOnly(r.handleMunicipalKakiAction))
	r.mux.HandleFunc("/api/v1/memory/search", r.handleMemorySearch)
	r.mux.HandleFunc("/api/v1/alerts/explain", r.handleAlertExplanation)
	r.mux.HandleFunc("/api/v1/ml/status", r.handleMlStatus)
	r.mux.HandleFunc("/api/v1/ml/retrain", r.adminOnly(r.handleMlRetrain))
	r.mux.HandleFunc("/api/v1/ml/predict", r.handleMlPredict)
	r.mux.HandleFunc("/api/v1/ml/query", r.handleMlQuery)
	r.mux.HandleFunc("/api/v1/bigquery/stream-sync", r.adminOnly(r.handleBigQueryStreamSync))
	r.mux.HandleFunc("/api/v1/bigquery/sync-status", r.handleBigQuerySyncStatus)
	r.mux.HandleFunc("/api/v1/bigquery/staged-samples", r.handleBigQueryStagedSamples)
	r.mux.HandleFunc("/api/v1/analytics/thailand-hec/crop-plots", r.handleThailandCropPlots)
	r.mux.HandleFunc("/api/v1/analytics/thailand-hec/recalculate", r.handleThailandHecRecalculate)
	r.mux.HandleFunc("/api/v1/analytics/thailand-hec/plots/", r.officialOnly(r.handleThailandPlotAction))
	r.mux.HandleFunc("/api/v1/analytics/sumatra-tiger/pastures", r.handleTigerPastures)
	r.mux.HandleFunc("/api/v1/analytics/sumatra-tiger/depredation/report", r.handleTigerDepredationReport)
	r.mux.HandleFunc("/api/v1/analytics/sumatra-tiger/recalculate", r.handleTigerRecalculate)
	r.mux.HandleFunc("/api/v1/analytics/sumatra-tiger/pastures/", r.officialOnly(r.handleTigerPastureAction))
	r.mux.HandleFunc("/api/v1/nlp/ingest/unstructured", r.handleUnstructuredNlpIngest)
	r.mux.HandleFunc("/api/v1/corroboration/streams", r.handleCorroborationStreams)
	r.mux.HandleFunc("/api/v1/corroboration/ingest", r.handleCorroborationIngest)
	r.mux.HandleFunc("/api/v1/corroboration/evaluate", r.handleCorroborationEvaluate)
	r.mux.HandleFunc("/api/v1/corroboration/poll-cycle", r.handleCorroborationPollCycle)

	// Cold-Start & Conservation NGO bulk ingestion routes (Issue #23)
	r.mux.HandleFunc("/api/v1/coldstart/gbif-firms/bootstrap", r.adminOnly(r.handleColdStartBootstrap))
	r.mux.HandleFunc("/api/v1/coldstart/status", r.handleColdStartStatus)
	r.mux.HandleFunc("/api/v1/ngo/patrol-logs/bulk-import", r.officialOnly(r.handleNgoBulkImport))
	r.mux.HandleFunc("/api/v1/ngo/patrol-logs/templates", r.handleNgoTemplates)

	// User role management route (Issue #46)
	r.mux.HandleFunc("/api/v1/admin/set-role", r.adminOnly(r.handleAdminSetRole))

	// Mount CoExist Shield AI Admin Web Portal endpoints (Issue #52, #53)
	if r.adminHandler != nil {
		r.adminHandler.RegisterRoutes(r.mux)
	}
}

func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "coexistshield-backend-cloudrun",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func (r *Router) handleRoot(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	r.handleHealth(w, req)
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

func (r *Router) handleGenerateForecastPlumeGeofences(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var forecastReq model.ForecastPlumeGeofencesRequest
	if err := json.NewDecoder(req.Body).Decode(&forecastReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON request: %v", err), http.StatusBadRequest)
		return
	}

	var forecastPoints []model.HourlyForecastPoint
	if len(forecastReq.Attractants) > 0 {
		refLat := forecastReq.Attractants[0].Latitude
		refLon := forecastReq.Attractants[0].Longitude
		ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
		defer cancel()
		pts, err := corroboration.FetchOpenMeteoHourlyForecast(ctx, nil, refLat, refLon, 24)
		if err == nil && len(pts) > 0 {
			forecastPoints = pts
		}
	}

	resp := r.geofenceEngine.GenerateForecastPlumeGeofences(forecastReq, forecastPoints)
	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) handleDeescalateGeofences(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	results := r.geofenceEngine.DeescalateActiveZones(time.Now().UTC())
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results":   results,
		"evaluated": len(results),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
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
	existingIDs := make(map[string]bool)
	for _, z := range zones {
		existingIDs[z.Id] = true
	}

	// Merge geofences from Admin Store (AI & manual geofences configured on Admin Portal)
	if r.adminHandler != nil && r.adminHandler.GetStore() != nil {
		adminGeos := r.adminHandler.GetStore().ListGeofences("")
		for _, ag := range adminGeos {
			if existingIDs[ag.ID] {
				continue
			}
			zone := adminGeofenceToActiveZone(ag)
			distKm := haversineDistanceKm(lat, lon, zone.Centroid.Latitude, zone.Centroid.Longitude)
			// Return if within radiusKm or if large regional overview is requested
			if radiusKm >= 1000.0 || distKm <= (radiusKm+zone.RadiusMeters/1000.0) {
				zones = append(zones, zone)
				existingIDs[ag.ID] = true
			}
		}
	}

	// Threat-Proximity Ranking & Quota Limiting (Mobile OS Geofence Limit Compliance)
	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	tierWeight := func(tier string) float64 {
		switch tier {
		case "TIER_3_DANGER", "TIER_3_CRITICAL":
			return 100.0
		case "TIER_2_WARNING", "TIER_2_ALERT":
			return 50.0
		case "TIER_1_NOTICE", "TIER_1_CAUTION":
			return 20.0
		default:
			return 5.0
		}
	}

	sort.Slice(zones, func(i, j int) bool {
		distI := haversineDistanceKm(lat, lon, zones[i].Centroid.Latitude, zones[i].Centroid.Longitude)
		distJ := haversineDistanceKm(lat, lon, zones[j].Centroid.Latitude, zones[j].Centroid.Longitude)
		scoreI := tierWeight(zones[i].Tier) / (1.0 + distI)
		scoreJ := tierWeight(zones[j].Tier) / (1.0 + distJ)
		return scoreI > scoreJ
	})

	if limit > 0 && len(zones) > limit {
		zones = zones[:limit]
	}

	if zones == nil {
		zones = []model.ActiveGeofenceZone{}
	}

	writeJSON(w, http.StatusOK, zones)
}

// handleGeofenceEvents establishes a Server-Sent Events (SSE) stream for real-time mobile geofence synchronization.
func (r *Router) handleGeofenceEvents(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	if r.geofenceEngine == nil || r.geofenceEngine.GetEventHub() == nil {
		http.Error(w, "Event hub unavailable", http.StatusServiceUnavailable)
		return
	}

	hub := r.geofenceEngine.GetEventHub()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := hub.Subscribe()
	defer hub.Unsubscribe(ch)

	// Send initial connection event
	initEvent := model.GeofenceEvent{
		Type:      "CONNECTED",
		Action:    "connected",
		Message:   "Connected to real-time geofence synchronization stream",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	initBytes, _ := json.Marshal(initEvent)
	fmt.Fprintf(w, "data: %s\n\n", initBytes)
	flusher.Flush()

	ctx := req.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			eventBytes, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", eventBytes)
			flusher.Flush()
		}
	}
}

func haversineDistanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180.0))*math.Cos(lat2*(math.Pi/180.0))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return 6371.0 * c
}

func adminGeofenceToActiveZone(g admin.AdminGeofence) model.ActiveGeofenceZone {
	tier := "TIER_2_WARNING"
	riskScore := 0.75
	switch g.ThreatLevel {
	case "TIER_3_CRITICAL", "TIER_3_DANGER":
		tier = "TIER_3_DANGER"
		riskScore = 0.95
	case "TIER_2_ALERT", "TIER_2_WARNING":
		tier = "TIER_2_WARNING"
		riskScore = 0.75
	case "TIER_1_CAUTION", "TIER_1_NOTICE":
		tier = "TIER_1_NOTICE"
		riskScore = 0.45
	default:
		tier = "SAFE"
		riskScore = 0.10
	}

	centroid := model.ActiveGeofencePoint{Latitude: 0, Longitude: 0}
	if g.CircleCenter != nil {
		centroid.Latitude = g.CircleCenter.Lat
		centroid.Longitude = g.CircleCenter.Lng
	} else if len(g.Coordinates) > 0 {
		var sumLat, sumLng float64
		for _, c := range g.Coordinates {
			sumLat += c.Lat
			sumLng += c.Lng
		}
		centroid.Latitude = sumLat / float64(len(g.Coordinates))
		centroid.Longitude = sumLng / float64(len(g.Coordinates))
	}

	radiusMeters := g.CircleRadiusMeters
	if radiusMeters <= 0 {
		if len(g.Coordinates) > 0 {
			var maxDist float64
			for _, c := range g.Coordinates {
				dist := haversineDistanceKm(centroid.Latitude, centroid.Longitude, c.Lat, c.Lng) * 1000.0
				if dist > maxDist {
					maxDist = dist
				}
			}
			radiusMeters = maxDist
		}
		if radiusMeters <= 0 {
			radiusMeters = 1000.0
		}
	}

	var coords []model.ActiveGeofencePoint
	if len(g.Coordinates) > 0 {
		coords = make([]model.ActiveGeofencePoint, len(g.Coordinates))
		for i, c := range g.Coordinates {
			coords[i] = model.ActiveGeofencePoint{Latitude: c.Lat, Longitude: c.Lng}
		}
	} else if centroid.Latitude != 0 || centroid.Longitude != 0 {
		// Generate 32 polygon vertices for circle representation
		coords = make([]model.ActiveGeofencePoint, 32)
		rDegLat := (radiusMeters / 6378137.0) * (180.0 / math.Pi)
		rDegLon := rDegLat
		cosLat := math.Cos(centroid.Latitude * math.Pi / 180.0)
		if cosLat > 0.0001 {
			rDegLon = rDegLat / cosLat
		}
		for i := 0; i < 32; i++ {
			theta := 2.0 * math.Pi * float64(i) / 32.0
			coords[i] = model.ActiveGeofencePoint{
				Latitude:  centroid.Latitude + rDegLat*math.Sin(theta),
				Longitude: centroid.Longitude + rDegLon*math.Cos(theta),
			}
		}
	}

	validFrom := g.CreatedAt.UTC().Format(time.RFC3339)
	validUntil := g.CreatedAt.Add(72 * time.Hour).UTC().Format(time.RFC3339)
	if g.ExpiresAt != nil {
		validUntil = g.ExpiresAt.UTC().Format(time.RFC3339)
	}

	targetSpecies := []string{g.TargetSpecies}
	if g.TargetSpecies == "" {
		targetSpecies = []string{"Wildlife"}
	}

	explanation := fmt.Sprintf("%s (%s): %s", g.Name, g.TargetSpecies, g.Reason)
	if g.Reason == "" {
		explanation = fmt.Sprintf("%s: Active barrier monitored by CoExist Shield AI", g.Name)
	}

	return model.ActiveGeofenceZone{
		Id:                   g.ID,
		Tier:                 tier,
		TargetSpecies:        targetSpecies,
		Coordinates:          coords,
		Centroid:             centroid,
		RadiusMeters:         radiusMeters,
		CausalExplanation:    explanation,
		PrimaryAttractantRef: g.ID,
		ActiveRiskScore:      riskScore,
		ValidFrom:            validFrom,
		ValidUntil:           validUntil,
		GeneratedAt:          validFrom,
	}
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
	explanation, recs, err := r.geminiClient.GenerateExplainableAlert(req.Context(), explainReq, analogs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Explanation synthesis error: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, model.CausalAlertExplanationResponse{
		Explanation:             explanation,
		HistoricalAnalogs:       analogs,
		ElevatedFromTier:        elevatedFrom,
		EffectiveTier:           effectiveTier,
		DispatchRecommendations: recs,
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

func (r *Router) handleBigQueryStreamSync(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var batch []ml.IncidentStreamingEvent
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil && len(bodyBytes) > 0 {
			var single ml.IncidentStreamingEvent
			if errSingle := json.Unmarshal(bodyBytes, &single); errSingle == nil && single.IncidentID != "" {
				batch = append(batch, single)
			} else {
				_ = json.Unmarshal(bodyBytes, &batch)
			}
		}
	}

	// If empty body or no valid events, return Bad Request
	if len(batch) == 0 {
		http.Error(w, "Batch payload is empty or invalid JSON", http.StatusBadRequest)
		return
	}

	count, err := r.bigqueryReplicator.StreamSyncBatch(batch)
	if err != nil {
		http.Error(w, fmt.Sprintf("BigQuery stream sync failed: %v", err), http.StatusInternalServerError)
		return
	}

	status := r.bigqueryReplicator.GetSyncStatus()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"syncedEventsCount": count,
		"syncStatus":        status,
		"stagedLakeSamples": len(r.bigqueryReplicator.GetStagedSamples()),
	})
}

func (r *Router) handleBigQuerySyncStatus(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	status := r.bigqueryReplicator.GetSyncStatus()
	writeJSON(w, http.StatusOK, status)
}

func (r *Router) handleBigQueryStagedSamples(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	samples := r.bigqueryReplicator.GetStagedSamples()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"samples": samples,
		"count":   len(samples),
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

func (r *Router) handleIncidents(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.handleGetIncidents(w, req)
	case http.MethodPost:
		r.handleSubmitIncident(w, req)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleGetIncidents(w http.ResponseWriter, req *http.Request) {
	category := req.URL.Query().Get("category")
	limit := 50
	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if r.incidentRepo != nil {
		incidents, err := r.incidentRepo.GetIncidents(req.Context(), limit)
		if err == nil && len(incidents) > 0 {
			var result []model.IncidentDto
			for _, inc := range incidents {
				if category == "" || strings.EqualFold(inc.Category, category) {
					result = append(result, inc)
				}
			}
			if result == nil {
				result = []model.IncidentDto{}
			}
			writeJSON(w, http.StatusOK, result)
			return
		}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []model.IncidentDto
	for _, inc := range r.incidents {
		if category == "" || strings.EqualFold(inc.Category, category) {
			result = append(result, inc)
			if len(result) >= limit {
				break
			}
		}
	}
	if result == nil {
		result = []model.IncidentDto{}
	}
	writeJSON(w, http.StatusOK, result)
}

func (r *Router) handleSubmitIncident(w http.ResponseWriter, req *http.Request) {
	var dto model.IncidentDto
	if err := json.NewDecoder(req.Body).Decode(&dto); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON request: %v", err), http.StatusBadRequest)
		return
	}

	if dto.Id == "" {
		dto.Id = fmt.Sprintf("inc_%d", time.Now().UnixNano())
	}
	if dto.Timestamp == "" {
		dto.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if dto.Region == "" {
		dto.Region = "JAPAN"
	}
	// Genuine Gemini 3.8 Flash triage validation if not already populated
	if dto.GeminiTriageResult == nil {
		var imageBase64 string
		if dto.PhotoBase64 != "" {
			imageBase64 = dto.PhotoBase64
		} else if len(dto.PhotoUrls) > 0 {
			// e.g. base64 embedded in first photo url or reference
			imageBase64 = dto.PhotoUrls[0]
		}

		if imageBase64 != "" && r.geminiClient != nil {
			triageReq := model.ForensicTriageRequest{
				ImageBase64:  imageBase64,
				CategoryHint: dto.Category,
				Latitude:     dto.Location.Latitude,
				Longitude:    dto.Location.Longitude,
			}
			triageResp, err := r.geminiClient.TriageFieldPhoto(req.Context(), triageReq)
			if err == nil && triageResp != nil {
				decay := ""
				fruitLoad := ""
				traceType := ""
				traumaType := ""
				if triageResp.Attributes != nil {
					if triageResp.Attributes.DecayState != nil {
						decay = *triageResp.Attributes.DecayState
					}
					fruitLoad = triageResp.Attributes.FruitLoadLevel
					if triageResp.Attributes.TraceType != nil {
						traceType = *triageResp.Attributes.TraceType
					}
					if triageResp.Attributes.TraumaType != nil {
						traumaType = *triageResp.Attributes.TraumaType
					}
				}
				dto.GeminiTriageResult = &model.GeminiTriageResultDto{
					DetectedLabel:  triageResp.DetectedSpecies,
					CommonName:     triageResp.CommonName,
					Confidence:     triageResp.Confidence,
					ForensicTags:   triageResp.SuggestedTags,
					DecayState:     decay,
					FruitLoadLevel: fruitLoad,
					TraceType:      traceType,
					TraumaType:     traumaType,
					Notes:          triageResp.Notes,
					ModelUsed:      gemini.GeminiModelVersion,
				}
				if triageResp.Confidence >= 0.70 {
					dto.VerificationStatus = "VERIFIED"
				} else {
					dto.VerificationStatus = "UNVERIFIED"
				}
			}
		}

		// Fallback for reports without photo evidence or when vision model unavailable:
		// Mark as UNVERIFIED with Confidence 0.0 rather than fabricating high confidence.
		if dto.GeminiTriageResult == nil {
			dto.VerificationStatus = "UNVERIFIED"
			dto.GeminiTriageResult = &model.GeminiTriageResultDto{
				DetectedLabel:   fmt.Sprintf("%s (%s)", dto.Species, dto.Category),
				CommonName:      dto.Species,
				Confidence:      0.0,
				ForensicTags:    []string{"unverified_crowdsource", strings.ToLower(dto.Category), strings.ToLower(dto.Species)},
				Notes:           "Pending multimodal vision triage: no verifiable field photo evidence submitted.",
				ModelUsed:       gemini.GeminiModelVersion,
				InferenceTimeMs: 0,
			}
		}
	} else if dto.VerificationStatus == "" {
		if dto.GeminiTriageResult.Confidence >= 0.70 {
			dto.VerificationStatus = "VERIFIED"
		} else {
			dto.VerificationStatus = "UNVERIFIED"
		}
	}

	if r.incidentRepo != nil {
		_ = r.incidentRepo.SaveIncident(req.Context(), dto)
	}

	r.mu.Lock()
	// If Roadkill, register roadkill carcass for predator lure mitigation
	if strings.EqualFold(dto.Category, "ROADKILL") {
		decay := "FRESH"
		if dto.GeminiTriageResult != nil && dto.GeminiTriageResult.DecayState != "" {
			decay = dto.GeminiTriageResult.DecayState
		}
		threatScore := computeRoadkillThreatScore(dto.Species, decay)
		rkRecord := model.RoadkillRecordDto{
			Id:                    fmt.Sprintf("rk_%d", time.Now().UnixNano()%100000),
			IncidentId:            dto.Id,
			Species:               dto.Species,
			Location:              dto.Location,
			DecayState:            decay,
			AttractantThreatScore: threatScore,
			ClearanceStatus:       "REPORTED",
			PhotoUrls:             dto.PhotoUrls,
			ReportedAt:            dto.Timestamp,
		}
		if r.roadkillRepo != nil {
			_ = r.roadkillRepo.SaveRoadkill(req.Context(), rkRecord)
		}
		r.roadkills = append([]model.RoadkillRecordDto{rkRecord}, r.roadkills...)
	}

	r.incidents = append([]model.IncidentDto{dto}, r.incidents...)
	r.mu.Unlock()

	// Recalculate dynamic geofence radius based on incident ground-truth
	if r.geofenceEngine != nil && dto.Location.Latitude != 0 && dto.Location.Longitude != 0 {
		r.geofenceEngine.Recalculate(model.GeofenceRecalculateRequest{
			TriggerEntityRef:      "incidents/" + dto.Id,
			TargetCenterLatitude:  dto.Location.Latitude,
			TargetCenterLongitude: dto.Location.Longitude,
			RadiusKm:              5.0,
		})
	}

	// Synchronize citizen/mobile incident into Admin store so AI Geofence recalculation correlates it
	if r.adminHandler != nil && r.adminHandler.GetStore() != nil && dto.Location.Latitude != 0 && dto.Location.Longitude != 0 {
		threat := "TIER_2_ALERT"
		if strings.EqualFold(dto.VerificationStatus, "VERIFIED") {
			threat = "TIER_3_CRITICAL"
		}
		r.adminHandler.GetStore().CreateWildlife(admin.AdminWildlifeIncident{
			ID:      dto.Id,
			Species: dto.Species,
			Location: admin.IncidentLocation{
				Lat:    dto.Location.Latitude,
				Lng:    dto.Location.Longitude,
				Region: dto.Region,
			},
			ThreatLevel: threat,
			Status:      dto.VerificationStatus,
			Reporter:    "Citizen Mobile User",
			Description: dto.Description,
			ReportedAt:  time.Now().UTC(),
		})
	}

	// Issue #20: Replicate incident event in real time to BigQuery GIS for feature enrichment & ML staging (<60s SLA)
	if r.bigqueryReplicator != nil && dto.Location.Latitude != 0 && dto.Location.Longitude != 0 {
		incursionVal := 0
		if dto.VerificationStatus == "VERIFIED" {
			incursionVal = 1
		}
		streamEvent := ml.IncidentStreamingEvent{
			EventID:           fmt.Sprintf("evt_%s", dto.Id),
			IncidentID:        dto.Id,
			Species:           dto.Species,
			Category:          dto.Category,
			Latitude:          dto.Location.Latitude,
			Longitude:         dto.Location.Longitude,
			Timestamp:         time.Now().UTC(),
			IsIncursion:       incursionVal,
			VerificationState: dto.VerificationStatus,
		}
		go func() {
			count, err := r.bigqueryReplicator.StreamSyncBatch([]ml.IncidentStreamingEvent{streamEvent})
			if err != nil {
				log.Printf("[BigQuery GIS] Stream sync error for incident %s: %v", dto.Id, err)
			} else {
				log.Printf("[BigQuery GIS] Successfully streamed %d event(s) to BigQuery GIS for incident %s", count, dto.Id)
			}
		}()
	}


	writeJSON(w, http.StatusCreated, dto)
}

func (r *Router) handleGetRoadkills(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.roadkillRepo != nil {
		roadkills, err := r.roadkillRepo.GetRoadkills(req.Context(), "")
		if err == nil && len(roadkills) > 0 {
			writeJSON(w, http.StatusOK, roadkills)
			return
		}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := r.roadkills
	if result == nil {
		result = []model.RoadkillRecordDto{}
	}
	writeJSON(w, http.StatusOK, result)
}

func (r *Router) handleRoadkillAction(w http.ResponseWriter, req *http.Request) {
	// Path pattern: /api/v1/roadkills/{id}/status
	path := strings.TrimPrefix(req.URL.Path, "/api/v1/roadkills/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "status" {
		http.NotFound(w, req)
		return
	}

	roadkillID := parts[0]
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var statusReq model.UpdateRoadkillStatusRequestDto
	if err := json.NewDecoder(req.Body).Decode(&statusReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if r.roadkillRepo != nil {
		updated, err := r.roadkillRepo.UpdateRoadkillStatus(req.Context(), roadkillID, statusReq.Status, "")
		if err == nil && updated != nil {
			if statusReq.AssignedOfficialId != "" {
				updated.AssignedOfficialId = statusReq.AssignedOfficialId
			}
			writeJSON(w, http.StatusOK, updated)
			return
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i, rk := range r.roadkills {
		if rk.Id == roadkillID {
			r.roadkills[i].ClearanceStatus = statusReq.Status
			if statusReq.AssignedOfficialId != "" {
				r.roadkills[i].AssignedOfficialId = statusReq.AssignedOfficialId
			}
			if statusReq.Status == "CLEARED" {
				r.roadkills[i].ClearedAt = time.Now().UTC().Format(time.RFC3339)
			}
			writeJSON(w, http.StatusOK, r.roadkills[i])
			return
		}
	}

	http.Error(w, fmt.Sprintf("Roadkill %s not found", roadkillID), http.StatusNotFound)
}

func (r *Router) initSeedData() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.incidents = []model.IncidentDto{
		{
			Id:       "inc_iwate_01",
			Category: "WILDLIFE_SIGHTING",
			Species:  "BEAR",
			Location: model.GeoLocationDto{
				Latitude:  39.7020,
				Longitude: 141.1520,
			},
			Timestamp:          "2026-10-14T01:30:00Z",
			Region:             "JAPAN",
			Description:        "Adult Asian Black Bear spotted feeding along edge of Takamatsu forest corridor.",
			ReporterId:         "usr_citizen_88",
			ReporterRole:       "CITIZEN",
			VerificationStatus: "VERIFIED",
			GeminiTriageResult: &model.GeminiTriageResultDto{
				DetectedLabel:   "Ursus thibetanus japonicus (Asian Black Bear)",
				CommonName:      "Asian Black Bear",
				Confidence:      0.95,
				ForensicTags:    []string{"active_sighting", "fall_hyperphagia", "iwate_corridor"},
				ModelUsed:       gemini.GeminiModelVersion,
				InferenceTimeMs: 148,
				Notes:           "High encounter probability confirmed in Takamatsu district.",
			},
		},
		{
			Id:       "inc_iwate_02",
			Category: "PHYSICAL_TRACE",
			Species:  "BEAR",
			Location: model.GeoLocationDto{
				Latitude:  39.7045,
				Longitude: 141.1490,
			},
			Timestamp:          "2026-10-14T02:15:00Z",
			Region:             "JAPAN",
			Description:        "Fresh bark claw marks and scat observed on abandoned persimmon tree trunk.",
			ReporterId:         "usr_ranger_04",
			ReporterRole:       "OFFICIAL",
			VerificationStatus: "VERIFIED",
			GeminiTriageResult: &model.GeminiTriageResultDto{
				DetectedLabel:   "Claw Mark Bark Stripping & Vegetative Scat",
				CommonName:      "Bear Foraging Trace",
				Confidence:      0.92,
				ForensicTags:    []string{"claw_mark", "persimmon_attractant", "scat_acorn"},
				TraceType:       "CLAW_MARK_BARK_STRIPPING",
				ModelUsed:       gemini.GeminiModelVersion,
				InferenceTimeMs: 135,
				Notes:           "Fresh tree marking within 120m of residential perimeter.",
			},
		},
		{
			Id:       "inc_iwate_03",
			Category: "ROADKILL",
			Species:  "WILD_BOAR",
			Location: model.GeoLocationDto{
				Latitude:  39.7120,
				Longitude: 141.1480,
			},
			Timestamp:          "2026-10-14T03:00:00Z",
			Region:             "JAPAN",
			Description:        "Decomposing wild boar carcass on Route 4 shoulder acting as predator attractant magnet.",
			ReporterId:         "usr_highway_patrol",
			ReporterRole:       "OFFICIAL",
			VerificationStatus: "VERIFIED",
			GeminiTriageResult: &model.GeminiTriageResultDto{
				DetectedLabel:   "Sus scrofa (Wild Boar)",
				CommonName:      "Wild Boar Carcass",
				Confidence:      0.94,
				ForensicTags:    []string{"carcass", "roadkill", "bloated_decay", "predator_lure"},
				DecayState:      "BLOATED",
				TraumaType:      "BLUNT_FORCE_VEHICULAR_IMPACT",
				ModelUsed:       gemini.GeminiModelVersion,
				InferenceTimeMs: 160,
				Notes:           "Bloated carcass emitting strong olfactory plume downwind.",
			},
		},
	}

	r.roadkills = []model.RoadkillRecordDto{
		{
			Id:                    "rk_4412",
			IncidentId:            "inc_iwate_03",
			Species:               "WILD_BOAR",
			Location:              model.GeoLocationDto{Latitude: 39.7120, Longitude: 141.1480},
			DecayState:            "BLOATED",
			AttractantThreatScore: 0.88,
			ClearanceStatus:       "REPORTED",
			ReportedAt:            "2026-10-14T03:00:00Z",
		},
		{
			Id:                    "rk_3901",
			IncidentId:            "inc_iwate_04",
			Species:               "DEER",
			Location:              model.GeoLocationDto{Latitude: 39.6410, Longitude: 141.9520},
			DecayState:            "DECOMPOSING",
			AttractantThreatScore: 0.75,
			ClearanceStatus:       "DISPATCHED",
			AssignedOfficialId:    "OFFICER-IWATE-42",
			ReportedAt:            "2026-10-13T18:00:00Z",
		},
		{
			Id:                    "rk_2105",
			IncidentId:            "inc_iwate_05",
			Species:               "RACCOON_DOG",
			Location:              model.GeoLocationDto{Latitude: 39.6800, Longitude: 141.1300},
			DecayState:            "SKELETAL",
			AttractantThreatScore: 0.20,
			ClearanceStatus:       "CLEARED",
			AssignedOfficialId:    "OFFICER-IWATE-18",
			ReportedAt:            "2026-10-12T09:00:00Z",
			ClearedAt:             "2026-10-12T14:30:00Z",
		},
	}

	if r.geofenceEngine != nil {
		r.geofenceEngine.LoadFixtureZones()
	}
	if r.coldstartEngine != nil {
		r.coldstartEngine.LoadFixtureMacroDatasets()
	}
}


func (r *Router) GetColdStartEngine() *coldstart.Engine {
	return r.coldstartEngine
}

func (r *Router) handleColdStartBootstrap(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	region := "ALL"
	if req.Body != nil {
		var body struct {
			Region string `json:"region"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err == nil && body.Region != "" {
			region = body.Region
		}
	}
	if q := req.URL.Query().Get("region"); q != "" {
		region = q
	}

	status := r.coldstartEngine.BootstrapDayOne(region)
	writeJSON(w, http.StatusOK, status)
}

func (r *Router) handleColdStartStatus(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	status := r.coldstartEngine.GetStatus()
	writeJSON(w, http.StatusOK, status)
}

func (r *Router) handleNgoBulkImport(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var importReq coldstart.NgoBulkImportRequest
	if err := json.NewDecoder(req.Body).Decode(&importReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON request body: %v", err), http.StatusBadRequest)
		return
	}

	resp, err := r.coldstartEngine.ImportPatrolLogs(importReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Bulk import failed: %v", err), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) handleNgoTemplates(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	templates := r.coldstartEngine.GetStandardTemplates()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"formats":   []string{"CSV", "GEOJSON", "GPX"},
		"sla":       "5 minutes",
	})
}

type AdminSetRoleRequest struct {
	UID  string `json:"uid"`
	Role string `json:"role"` // CITIZEN, OFFICIAL, ADMIN
}

func (r *Router) handleAdminSetRole(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var roleReq AdminSetRoleRequest
	if err := json.NewDecoder(req.Body).Decode(&roleReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON request: %v", err), http.StatusBadRequest)
		return
	}

	role := strings.ToUpper(strings.TrimSpace(roleReq.Role))
	if role != "CITIZEN" && role != "OFFICIAL" && role != "ADMIN" {
		http.Error(w, "Invalid role: must be CITIZEN, OFFICIAL, or ADMIN", http.StatusBadRequest)
		return
	}
	if roleReq.UID == "" {
		http.Error(w, "Missing required uid", http.StatusBadRequest)
		return
	}

	r.mu.Lock()
	if r.userRoles == nil {
		r.userRoles = make(map[string]string)
	}
	r.userRoles[roleReq.UID] = role
	r.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"uid":       roleReq.UID,
		"role":      role,
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
