package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"ai.coexistshield/backend/internal/model"
)

// GeminiModelVersion strictly defines Gemini 3.8 Flash per hackathon / architectural mandate.
const GeminiModelVersion = "gemini-3.8-flash"
const DefaultGeminiEndpoint = "https://generativelanguage.googleapis.com/v1beta"

var ErrApiKeyMissing = fmt.Errorf("gemini API key is not configured and GEMINI_MOCK_ENABLED is false")

type Config struct {
	ApiKey      string
	Model       string
	Endpoint    string
	MockEnabled bool
	Timeout     time.Duration
}

type Client struct {
	apiKey      string
	model       string
	endpoint    string
	mockEnabled bool
	httpClient  *http.Client
}

func NewClientWithConfig(cfg Config) *Client {
	modelName := cfg.Model
	if modelName == "" {
		modelName = GeminiModelVersion
	}
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = DefaultGeminiEndpoint
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	return &Client{
		apiKey:      cfg.ApiKey,
		model:       modelName,
		endpoint:    endpoint,
		mockEnabled: cfg.MockEnabled,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func isMockEnabledFromEnv() bool {
	val := strings.ToLower(strings.TrimSpace(os.Getenv("GEMINI_MOCK_ENABLED")))
	if val == "true" || val == "1" || val == "yes" {
		return true
	}
	valMock := strings.ToLower(strings.TrimSpace(os.Getenv("GEMINI_MOCK")))
	return valMock == "true" || valMock == "1" || valMock == "yes"
}

func NewClient() *Client {
	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = GeminiModelVersion
	}
	endpoint := os.Getenv("GEMINI_API_ENDPOINT")
	if endpoint == "" {
		endpoint = DefaultGeminiEndpoint
	}

	return NewClientWithConfig(Config{
		ApiKey:      os.Getenv("GEMINI_API_KEY"),
		Model:       modelName,
		Endpoint:    endpoint,
		MockEnabled: isMockEnabledFromEnv(),
		Timeout:     15 * time.Second,
	})
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
	Temperature      float64 `json:"temperature"`
}

type geminiRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiCandidate struct {
	Content struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

// executeGeminiCall handles HTTP POST with exponential backoff for transient 429/5xx status codes
func (c *Client) executeGeminiCall(ctx context.Context, gReq geminiRequest) (string, error) {
	bodyBytes, err := json.Marshal(gReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal gemini request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.endpoint, c.model, c.apiKey)

	maxRetries := 2
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
		if err != nil {
			return "", err
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			continue
		}

		respBytes, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var gResp geminiResponse
			if err := json.Unmarshal(respBytes, &gResp); err != nil {
				return "", fmt.Errorf("failed to decode gemini response json: %w", err)
			}
			if len(gResp.Candidates) == 0 || len(gResp.Candidates[0].Content.Parts) == 0 {
				return "", fmt.Errorf("gemini response candidates are empty")
			}
			return gResp.Candidates[0].Content.Parts[0].Text, nil
		}

		// Transient retriable statuses: 429, 500, 502, 503, 504
		if resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode >= 500 && resp.StatusCode <= 504) {
			lastErr = fmt.Errorf("gemini api error status %d: %s", resp.StatusCode, string(respBytes))
			continue
		}

		// Non-retriable 4xx client errors
		return "", fmt.Errorf("gemini api error status %d: %s", resp.StatusCode, string(respBytes))
	}

	return "", fmt.Errorf("gemini api call failed after %d retries: %w", maxRetries, lastErr)
}

// TriageFieldPhoto performs multimodal forensic analysis using Gemini 3.8 Flash.
// Evaluates wildlife species, carcass decay states, trauma, and physical traces (claw marks, pugmarks, scat).
func (c *Client) TriageFieldPhoto(ctx context.Context, req model.ForensicTriageRequest) (*model.ForensicTriageResponse, error) {
	if c.apiKey == "" {
		if c.mockEnabled {
			return mockTriage(req, "mock"), nil
		}
		return nil, ErrApiKeyMissing
	}

	rawBase64 := req.ImageBase64
	mimeType := "image/jpeg"
	if strings.Contains(rawBase64, ",") {
		parts := strings.SplitN(rawBase64, ",", 2)
		if strings.Contains(parts[0], "image/png") {
			mimeType = "image/png"
		}
		rawBase64 = parts[1]
	}

	prompt := fmt.Sprintf(`You are an expert wildlife biologist, field forensic tracker, and veterinary forensic pathologist.
Analyze this field photograph taken at latitude: %.4f, longitude: %.4f.
Category Hint: %s.

Perform multimodal forensic triage using Gemini 3.8 Flash. Output strict JSON adhering to schema:
{
  "detectedSpecies": string (scientific binomial, e.g. "Ursus thibetanus", "Panthera tigris sumatrae", "Elephas maximus", "Sus scrofa", "Diospyros kaki"),
  "commonName": string,
  "confidence": float (0.0 to 1.0),
  "attributes": {
    "fruitLoadLevel": "SPARSE" | "MODERATE" | "ABUNDANT" (if tree/crop),
    "ripenessStage": "UNRIPE_GREEN" | "RIPE" | "OVERRIPE_FALLING" (if tree/crop),
    "decayState": "FRESH" | "BLOATED" | "DECOMPOSING" | "SKELETAL" (if carcass/kill),
    "traumaType": "BLUNT_FORCE_VEHICULAR" | "CANINE_PUNCTURE" | "CLAW_LACERATION" | "GUNSHOT" | "NATURAL_DEATH" (if injury/kill),
    "traceType": "CLAW_MARK_BARK_STRIPPING" | "PUGMARK_SPREAD" | "SCAT_VEGETATIVE_ACORN" | "HAIR_RUB_RESIDUE" | "FENCE_BREACH" (if trace),
    "traceAgeHours": float (estimated age in hours)
  },
  "suggestedTags": array of strings,
  "preliminaryThreatIndex": float (0.0 to 1.0),
  "notes": brief one-sentence explanation
}`, req.Latitude, req.Longitude, req.CategoryHint)

	gReq := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{
						InlineData: &geminiInlineData{
							MimeType: mimeType,
							Data:     rawBase64,
						},
					},
					{
						Text: prompt,
					},
				},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			ResponseMimeType: "application/json",
			Temperature:      0.2,
		},
	}

	text, err := c.executeGeminiCall(ctx, gReq)
	if err != nil {
		if c.mockEnabled {
			return mockTriage(req, "mock_fallback"), nil
		}
		return nil, err
	}

	var triageResp model.ForensicTriageResponse
	if err := json.Unmarshal([]byte(text), &triageResp); err != nil {
		if c.mockEnabled {
			return mockTriage(req, "mock_fallback"), nil
		}
		return nil, fmt.Errorf("failed to parse structured gemini json: %w", err)
	}

	triageResp.Source = "gemini"
	return &triageResp, nil
}

func mockTriage(req model.ForensicTriageRequest, source string) *model.ForensicTriageResponse {
	cat := strings.ToUpper(req.CategoryHint)
	var resp *model.ForensicTriageResponse

	if cat == "ATTRACTANT_FRUIT_TREE" {
		resp = &model.ForensicTriageResponse{
			DetectedSpecies: "Diospyros kaki",
			CommonName:      "Persimmon Tree",
			Confidence:      0.96,
			Attributes: &model.TriageAttributes{
				FruitLoadLevel: "ABUNDANT",
				RipenessStage:  "OVERRIPE_FALLING",
			},
			SuggestedTags:          []string{"abundant_fruit", "overripe_falling", "high_calorie_lure"},
			PreliminaryThreatIndex: 0.88,
			Notes:                  "High bear attractant hazard identified via Gemini 3.8 Flash.",
		}
	} else if cat == "ROADKILL" {
		fresh := "FRESH"
		blunt := "BLUNT_FORCE_VEHICULAR"
		resp = &model.ForensicTriageResponse{
			DetectedSpecies: "Sus scrofa",
			CommonName:      "Wild Boar",
			Confidence:      0.92,
			Attributes: &model.TriageAttributes{
				DecayState: &fresh,
				TraumaType: &blunt,
			},
			SuggestedTags:          []string{"fresh_carcass", "predator_lure", "traffic_hazard"},
			PreliminaryThreatIndex: 0.82,
			Notes:                  "Fresh roadside carcass detected; creates immediate apex predator lure.",
		}
	} else if cat == "PHYSICAL_TRACE" {
		traceType := "CLAW_MARK_BARK_STRIPPING"
		traceAge := 3.5
		resp = &model.ForensicTriageResponse{
			DetectedSpecies: "Ursus thibetanus",
			CommonName:      "Asian Black Bear Claw Marks",
			Confidence:      0.95,
			Attributes: &model.TriageAttributes{
				TraceType:     &traceType,
				TraceAgeHours: &traceAge,
			},
			SuggestedTags:          []string{"fresh_claw_marks", "conifer_bark_stripped", "territorial_marking"},
			PreliminaryThreatIndex: 0.85,
			Notes:                  "Fresh Asian Black Bear claw marks on Japanese cedar; estimated < 4 hours old.",
		}
	} else if cat == "LIVESTOCK_DEPREDATION" {
		fresh := "FRESH"
		canine := "CANINE_PUNCTURE"
		resp = &model.ForensicTriageResponse{
			DetectedSpecies: "Panthera tigris sumatrae",
			CommonName:      "Sumatra Tiger Depredation",
			Confidence:      0.97,
			Attributes: &model.TriageAttributes{
				DecayState: &fresh,
				TraumaType: &canine,
			},
			SuggestedTags:          []string{"fresh_kill_site", "canine_puncture", "ambush_cache"},
			PreliminaryThreatIndex: 0.96,
			Notes:                  "Verified fresh tiger depredation; apex predator caching behavior detected.",
		}
	} else {
		// Default wildlife sighting
		resp = &model.ForensicTriageResponse{
			DetectedSpecies:        "Ursus thibetanus",
			CommonName:             "Asian Black Bear",
			Confidence:             0.94,
			SuggestedTags:          []string{"wildlife_sighting", "urban_interface"},
			PreliminaryThreatIndex: 0.78,
			Notes:                  "Verified wildlife sighting in suburban fringe zone via Gemini 3.8 Flash.",
		}
	}

	resp.Source = source
	return resp
}

// ExtractIncidentEntities parses unstructured municipal emergency broadcasts, social posts (X, Bluesky),
// and citizen reports to extract structured spatio-temporal incident entities using Gemini 3.8 Flash.
func (c *Client) ExtractIncidentEntities(ctx context.Context, text string, sourceChannel string, regionHint string) (*model.ExtractedIncidentEntity, error) {
	if c.apiKey == "" {
		if c.mockEnabled {
			return mockExtractEntities(text, sourceChannel, regionHint, "mock"), nil
		}
		return nil, ErrApiKeyMissing
	}

	prompt := fmt.Sprintf(`You are an expert NLP extraction engine for wildlife early-warning and human-wildlife coexistence platforms.
Analyze this raw unstructured incident report from channel: %s (Region Hint: %s).
Raw Text:
"""
%s
"""

Use Gemini 3.8 Flash to extract structured spatio-temporal incident entities.
Output strict JSON with schema:
{
  "entityId": string (uuid or timestamp hash),
  "species": "BEAR" | "ELEPHANT" | "TIGER" | "WILD_BOAR" | "DEER" | "UNKNOWN",
  "scientificName": string (e.g. "Ursus thibetanus", "Elephas maximus", "Panthera tigris sumatrae"),
  "commonName": string,
  "incidentType": "LIVE_SIGHTING" | "PHYSICAL_TRACE" | "CROP_RAID" | "DEPREDATION" | "ROADKILL",
  "locationDescription": string (exact landmark, district, or school name mentioned),
  "estimatedLatitude": float (inferred from location, e.g. 39.702 or 13.521 or 0.825),
  "estimatedLongitude": float (inferred from location, e.g. 141.152 or 101.452 or 101.530),
  "individualCount": int (e.g. 1, 3),
  "attractantContext": string (e.g. "Abandoned persimmon tree", "Sugarcane plantation border", "Uncollected garbage spot"),
  "urgencyLevel": "LOW" | "MEDIUM" | "HIGH" | "CRITICAL",
  "confidence": float (0.0 to 1.0),
  "corroborationTags": array of strings (keywords for cross-sensor correlation),
  "notes": brief one-sentence summary
}`, sourceChannel, regionHint, text)

	gReq := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{
						Text: prompt,
					},
				},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			ResponseMimeType: "application/json",
			Temperature:      0.1,
		},
	}

	responseText, err := c.executeGeminiCall(ctx, gReq)
	if err != nil {
		if c.mockEnabled {
			return mockExtractEntities(text, sourceChannel, regionHint, "mock_fallback"), nil
		}
		return nil, err
	}

	var entity model.ExtractedIncidentEntity
	if err := json.Unmarshal([]byte(responseText), &entity); err != nil {
		if c.mockEnabled {
			return mockExtractEntities(text, sourceChannel, regionHint, "mock_fallback"), nil
		}
		return nil, fmt.Errorf("failed to parse structured gemini json: %w", err)
	}

	entity.RawText = text
	entity.SourceChannel = sourceChannel
	entity.InferenceProvider = "gemini"
	if entity.ExtractedTimestamp == "" {
		entity.ExtractedTimestamp = time.Now().UTC().Format(time.RFC3339)
	}

	return &entity, nil
}

func mockExtractEntities(text string, sourceChannel string, regionHint string, source string) *model.ExtractedIncidentEntity {
	lower := strings.ToLower(text)
	nowStr := time.Now().UTC().Format(time.RFC3339)

	// Check Elephant / Thailand
	if strings.Contains(text, "ช้าง") || strings.Contains(lower, "elephant") || strings.Contains(lower, "chachoengsao") || regionHint == "THAILAND" {
		count := 1
		if strings.Contains(text, "3") || strings.Contains(lower, "three") {
			count = 3
		}
		entity := &model.ExtractedIncidentEntity{
			EntityId:            fmt.Sprintf("nlp-ent-%d", time.Now().UnixNano()%1000000),
			SourceChannel:       sourceChannel,
			RawText:             text,
			Species:             "ELEPHANT",
			ScientificName:      "Elephas maximus",
			CommonName:          "Asian Elephant",
			IncidentType:        "CROP_RAID",
			LocationDescription: "Khao Ang Rue Nai forest border, Nong Khla sugarcane plot",
			EstimatedLatitude:   13.5214,
			EstimatedLongitude:  101.4528,
			IndividualCount:     count,
			AttractantContext:   "Ripe sugarcane & pineapple plantation border",
			UrgencyLevel:        "HIGH",
			Confidence:          0.94,
			ExtractedTimestamp:  nowStr,
			CorroborationTags:   []string{"elephant_herd", "sugarcane_raid", "khao_ang_rue_nai"},
			Notes:               "Gemini 3.8 Flash extracted herd movement toward harvestable crops.",
			InferenceProvider:   source,
		}
		return entity
	}

	// Check Tiger / Indonesia / Riau
	if strings.Contains(lower, "tiger") || strings.Contains(lower, "harimau") || strings.Contains(lower, "siak") || strings.Contains(lower, "riau") || regionHint == "INDONESIA" {
		isTrace := strings.Contains(lower, "pugmark") || strings.Contains(lower, "trace") || strings.Contains(lower, "jejak")
		incType := "LIVE_SIGHTING"
		if isTrace {
			incType = "PHYSICAL_TRACE"
		} else if strings.Contains(lower, "attack") || strings.Contains(lower, "cattle") || strings.Contains(lower, "depredation") {
			incType = "DEPREDATION"
		}

		return &model.ExtractedIncidentEntity{
			EntityId:            fmt.Sprintf("nlp-ent-%d", time.Now().UnixNano()%1000000),
			SourceChannel:       sourceChannel,
			RawText:             text,
			Species:             "TIGER",
			ScientificName:      "Panthera tigris sumatrae",
			CommonName:          "Sumatra Tiger",
			IncidentType:        incType,
			LocationDescription: "Siak River riparian buffer, Block 14 Oil Palm Plantation",
			EstimatedLatitude:   0.8251,
			EstimatedLongitude:  101.5302,
			IndividualCount:     1,
			AttractantContext:   "Riparian secondary scrub bordering night corral",
			UrgencyLevel:        "CRITICAL",
			Confidence:          0.96,
			ExtractedTimestamp:  nowStr,
			CorroborationTags:   []string{"sumatra_tiger", "oil_palm_block", "siak_river", "fresh_pugmark"},
			Notes:               "Gemini 3.8 Flash parsed tiger presence in high-risk plantation interface.",
			InferenceProvider:   source,
		}
	}

	// Check Bear / Japan (default / Japanese text)
	isTrace := strings.Contains(text, "痕跡") || strings.Contains(text, "足跡") || strings.Contains(text, "爪痕") || strings.Contains(lower, "claw") || strings.Contains(lower, "paw")
	incType := "LIVE_SIGHTING"
	if isTrace {
		incType = "PHYSICAL_TRACE"
	}

	urgency := "MEDIUM"
	if strings.Contains(text, "小学校") || strings.Contains(text, "児童") || strings.Contains(text, "通学路") || strings.Contains(lower, "school") {
		urgency = "HIGH"
	}
	if strings.Contains(text, "人身") || strings.Contains(text, "負傷") || strings.Contains(text, "襲撃") || strings.Contains(lower, "attack") {
		urgency = "CRITICAL"
		incType = "DEPREDATION"
	}

	loc := "盛岡市松ヶ枝町付近 (松ヶ枝小学校南300m)"
	if locMatch := regexp.MustCompile(`([^\s]+市[^\s]+町?|[^\s]+付近)`).FindString(text); locMatch != "" {
		loc = locMatch
	}

	return &model.ExtractedIncidentEntity{
		EntityId:            fmt.Sprintf("nlp-ent-%d", time.Now().UnixNano()%1000000),
		SourceChannel:       sourceChannel,
		RawText:             text,
		Species:             "BEAR",
		ScientificName:      "Ursus thibetanus",
		CommonName:          "Asian Black Bear",
		IncidentType:        incType,
		LocationDescription: loc,
		EstimatedLatitude:   39.7020,
		EstimatedLongitude:  141.1520,
		IndividualCount:     1,
		AttractantContext:   "Unpicked persimmon tree near residential hedge",
		UrgencyLevel:        urgency,
		Confidence:          0.95,
		ExtractedTimestamp:  nowStr,
		CorroborationTags:   []string{"bear_sighting", "kaki_attractant", "school_route"},
		Notes:               "Gemini 3.8 Flash causal NLP extracted bear incident from municipal stream.",
		InferenceProvider:   source,
	}
}

// GenerateExplainableAlert synthesizes causal weather, attractants, multi-sensor corroboration, and episodic analogs
// into an urgent, transparent natural language justification displayed on the Main HUD, along with prioritized
// municipal dispatch recommendations using Gemini 3.8 Flash.
func (c *Client) GenerateExplainableAlert(ctx context.Context, req model.CausalAlertExplanationRequest, analogs []model.IncidentMemoryRecord) (string, []model.DispatchRecommendation, error) {
	if c.apiKey == "" {
		if c.mockEnabled {
			expl, recs := mockExplanation(req, analogs)
			return expl, recs, nil
		}
		return "", nil, ErrApiKeyMissing
	}

	var analogContext string
	if len(analogs) > 0 {
		analogContext = fmt.Sprintf("Historical Analogs: %d past incident(s) under similar causal signature. Primary analog: %s (Outcome: %s)", len(analogs), analogs[0].Narrative, analogs[0].Outcome)
	}

	prompt := fmt.Sprintf(`You are an autonomous wildlife protection AI powered by Gemini 3.8 Flash.
Synthesize a single, urgent but calm sentence in English (max 25 words) explaining why this specific risk tier was triggered based on causal drivers, multi-sensor corroboration, and historical memory.

Target Tier: %s
Species: %s
Weather: %+v
Nearby Attractants: %+v
%s

Direct output: Only the single sentence without quotes or preamble.`, req.Tier, req.TargetSpecies, req.Weather, req.NearbyAttractants, analogContext)

	gReq := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{
						Text: prompt,
					},
				},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			Temperature: 0.2,
		},
	}

	text, err := c.executeGeminiCall(ctx, gReq)
	if err != nil {
		if c.mockEnabled {
			expl, recs := mockExplanation(req, analogs)
			return expl, recs, nil
		}
		return "", nil, err
	}

	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		if c.mockEnabled {
			expl, recs := mockExplanation(req, analogs)
			return expl, recs, nil
		}
		return "", nil, fmt.Errorf("gemini returned empty explanation text")
	}

	recs := buildDispatchRecommendations(req, analogs)
	return trimmed, recs, nil
}

func mockExplanation(req model.CausalAlertExplanationRequest, analogs []model.IncidentMemoryRecord) (string, []model.DispatchRecommendation) {
	recs := buildDispatchRecommendations(req, analogs)

	if len(analogs) > 0 {
		return fmt.Sprintf("%s: Conditions match %d historical %s incident(s) with %s attractants; elevated caution advised via Gemini 3.8 Flash.",
			req.Tier, len(analogs), strings.ToLower(req.TargetSpecies), strings.ToLower(analogs[0].Category)), recs
	}

	if req.Tier == "TIER_3_DANGER" {
		return fmt.Sprintf("IMMEDIATE DANGER: Fresh %s kill-site identified within striking corridor; probability of predator return within 72h is critical.",
			strings.ToLower(req.TargetSpecies)), recs
	} else if req.Tier == "TIER_2_WARNING" {
		return fmt.Sprintf("Tier 2 Warning: New Moon darkness combined with prevailing wind toward unharvested %s attractants elevates intrusion risk.",
			strings.ToLower(req.TargetSpecies)), recs
	}
	return "Clear Zone: Continuous causal sensors indicate safe passage within regional radius.", recs
}

func buildDispatchRecommendations(req model.CausalAlertExplanationRequest, analogs []model.IncidentMemoryRecord) []model.DispatchRecommendation {
	spec := strings.ToUpper(req.TargetSpecies)
	region := strings.ToUpper(req.Region)

	if spec == "ELEPHANT" || region == "THAILAND" {
		return []model.DispatchRecommendation{
			{
				ActionType:      "BUFFER_HARVEST_MONITORING",
				Priority:        "IMMEDIATE",
				TargetEntityRef: "plot_th_01",
				Location:        "Khao Ang Rue Nai Border (Nong Khla Sugarcane Plot)",
				Justification:   "Ripening sugarcane aroma plume aligns with evening migration corridor. Deploy village push squad.",
			},
			{
				ActionType:      "ARM_ACOUSTIC_DEFLECTION",
				Priority:        "HIGH",
				TargetEntityRef: "flare_grid_04",
				Location:        "Corridor Passage Alpha (Watchtower 2)",
				Justification:   "Arm non-lethal acoustic deterrent arrays to safely steer herd away from human settlements.",
			},
		}
	}

	if spec == "TIGER" || region == "INDONESIA" {
		return []model.DispatchRecommendation{
			{
				ActionType:      "SECURE_KILL_SITE_CARCASS",
				Priority:        "IMMEDIATE",
				TargetEntityRef: "ls_sumatra_01",
				Location:        "Block 14 Pasture Buffer & Riparian Scrub",
				Justification:   "Carcass cached < 48 hours. Remove or incinerate remains under armed ranger escort to prevent fatal return encounter.",
			},
			{
				ActionType:      "ENFORCE_NIGHT_CORRALS",
				Priority:        "HIGH",
				TargetEntityRef: "enclosures_division_2",
				Location:        "Desa Bukit Kesuma Cattle Enclosures",
				Justification:   "Mandate all livestock secured in predator-proof night corrals during critical 72-hour return window.",
			},
		}
	}

	// Default: Bear / Japan
	return []model.DispatchRecommendation{
		{
			ActionType:      "FELL_PERSIMMON_TREE",
			Priority:        "IMMEDIATE",
			TargetEntityRef: "kt_1120",
			Location:        "盛岡市高松3丁目 (松ヶ枝小学校通学路 120m)",
			Justification:   "Overripe persimmon tree directly upwind of residential perimeter matches 3 historical bear intrusions. Priority felling eliminates hyperphagia lure.",
		},
		{
			ActionType:      "CLEAR_ROADKILL_CARCASS",
			Priority:        "IMMEDIATE",
			TargetEntityRef: "rk_4412",
			Location:        "Route 4 Takamatsu Bypass Shoulder",
			Justification:   "Bloated wild boar carcass emits downwind olfactory plume acting as predator magnet toward urban fringe.",
		},
		{
			ActionType:      "PATROL_CORRIDOR",
			Priority:        "SCHEDULED",
			TargetEntityRef: "patrol_squad_03",
			Location:        "Takamatsu Forest Fringe & School Corridor",
			Justification:   "Deploy municipal wildlife control squad with non-lethal acoustic deterrents during peak dawn/dusk hours.",
		},
	}
}
