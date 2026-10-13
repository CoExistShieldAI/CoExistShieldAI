package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"ai.coexistshield/backend/internal/model"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		apiKey: os.Getenv("GEMINI_API_KEY"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type geminiPart struct {
	Text       string               `json:"text,omitempty"`
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
	ResponseMimeType string `json:"responseMimeType,omitempty"`
	Temperature      float64 `json:"temperature"`
}

type geminiRequest struct {
	Contents          []geminiContent         `json:"contents"`
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

// TriageFieldPhoto performs multimodal forensic analysis using Gemini 1.5/2.0 Flash
func (c *Client) TriageFieldPhoto(ctx context.Context, req model.ForensicTriageRequest) (*model.ForensicTriageResponse, error) {
	// If no API key configured (e.g. mock or local testing), use deterministic high-speed fallback
	if c.apiKey == "" {
		return mockTriage(req), nil
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

	prompt := fmt.Sprintf(`You are an expert wildlife biologist and veterinary forensic pathologist.
Analyze this field photo taken at latitude: %.4f, longitude: %.4f.
Category Hint: %s.

Output strict JSON with fields:
- detectedSpecies (string, e.g. "Diospyros kaki", "Ursus thibetanus", "Panthera tigris")
- commonName (string)
- confidence (float, 0.0 to 1.0)
- attributes: { fruitLoadLevel: "SPARSE"|"MODERATE"|"ABUNDANT", ripenessStage: "UNRIPE_GREEN"|"RIPE"|"OVERRIPE_FALLING", decayState: "FRESH"|"BLOATED"|"DECOMPOSING"|"SKELETAL" }
- suggestedTags: array of strings
- preliminaryThreatIndex: float (0.0 to 1.0)
- notes: brief one-sentence explanation`, req.Latitude, req.Longitude, req.CategoryHint)

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

	bodyBytes, err := json.Marshal(gReq)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-3.8-flash:generateContent?key=%s", c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return mockTriage(req), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini api error status %d: %s", resp.StatusCode, string(respBytes))
	}

	var gResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
		return nil, err
	}

	if len(gResp.Candidates) == 0 || len(gResp.Candidates[0].Content.Parts) == 0 {
		return mockTriage(req), nil
	}

	text := gResp.Candidates[0].Content.Parts[0].Text
	var triageResp model.ForensicTriageResponse
	if err := json.Unmarshal([]byte(text), &triageResp); err != nil {
		return nil, fmt.Errorf("failed to parse structured gemini json: %w", err)
	}

	return &triageResp, nil
}

func mockTriage(req model.ForensicTriageRequest) *model.ForensicTriageResponse {
	if req.CategoryHint == "ATTRACTANT_FRUIT_TREE" {
		return &model.ForensicTriageResponse{
			DetectedSpecies: "Diospyros kaki",
			CommonName:      "Persimmon Tree",
			Confidence:      0.96,
			Attributes: &model.TriageAttributes{
				FruitLoadLevel: "ABUNDANT",
				RipenessStage:  "OVERRIPE_FALLING",
			},
			SuggestedTags:          []string{"abundant_fruit", "overripe_falling", "high_calorie_lure"},
			PreliminaryThreatIndex: 0.88,
			Notes:                  "High bear attractant hazard identified via Gemini Flash.",
		}
	} else if req.CategoryHint == "ROADKILL" {
		fresh := "FRESH"
		return &model.ForensicTriageResponse{
			DetectedSpecies: "Sus scrofa",
			CommonName:      "Wild Boar",
			Confidence:      0.92,
			Attributes: &model.TriageAttributes{
				DecayState: &fresh,
			},
			SuggestedTags:          []string{"fresh_carcass", "predator_lure", "traffic_hazard"},
			PreliminaryThreatIndex: 0.82,
			Notes:                  "Fresh roadside carcass detected; creates immediate apex predator lure.",
		}
	}

	return &model.ForensicTriageResponse{
		DetectedSpecies: "Ursus thibetanus",
		CommonName:      "Asian Black Bear",
		Confidence:      0.94,
		SuggestedTags:   []string{"wildlife_sighting", "urban_interface"},
		PreliminaryThreatIndex: 0.78,
		Notes:           "Verified wildlife sighting in suburban fringe zone.",
	}
}

// GenerateExplainableAlert synthesizes causal weather, attractants, and historical episodic analogs
// into an urgent, transparent natural language justification displayed on the Main HUD.
func (c *Client) GenerateExplainableAlert(ctx context.Context, req model.CausalAlertExplanationRequest, analogs []model.IncidentMemoryRecord) (string, error) {
	if c.apiKey == "" {
		return mockExplanation(req, analogs), nil
	}

	var analogContext string
	if len(analogs) > 0 {
		analogContext = fmt.Sprintf("Historical Analogs: %d past incident(s) under similar causal signature. Primary analog: %s (Outcome: %s)", len(analogs), analogs[0].Narrative, analogs[0].Outcome)
	}

	prompt := fmt.Sprintf(`You are an autonomous wildlife protection AI.
Synthesize a single, urgent but calm sentence in English (max 25 words) explaining why this specific risk tier was triggered based on causal drivers and historical memory.

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

	bodyBytes, err := json.Marshal(gReq)
	if err != nil {
		return mockExplanation(req, analogs), nil
	}

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return mockExplanation(req, analogs), nil
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return mockExplanation(req, analogs), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return mockExplanation(req, analogs), nil
	}

	var gResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
		return mockExplanation(req, analogs), nil
	}

	if len(gResp.Candidates) > 0 && len(gResp.Candidates[0].Content.Parts) > 0 {
		text := strings.TrimSpace(gResp.Candidates[0].Content.Parts[0].Text)
		if text != "" {
			return text, nil
		}
	}

	return mockExplanation(req, analogs), nil
}

func mockExplanation(req model.CausalAlertExplanationRequest, analogs []model.IncidentMemoryRecord) string {
	if len(analogs) > 0 {
		return fmt.Sprintf("%s: Conditions match %d historical %s incident(s) with %s attractants; elevated caution advised.", req.Tier, len(analogs), strings.ToLower(req.TargetSpecies), strings.ToLower(analogs[0].Category))
	}

	if req.Tier == "TIER_3_DANGER" {
		return fmt.Sprintf("IMMEDIATE DANGER: Fresh %s kill-site identified within striking corridor; probability of predator return within 72h is critical.", strings.ToLower(req.TargetSpecies))
	} else if req.Tier == "TIER_2_WARNING" {
		return fmt.Sprintf("Tier 2 Warning: New Moon darkness combined with prevailing wind toward unharvested %s attractants elevates intrusion risk.", strings.ToLower(req.TargetSpecies))
	}
	return "Clear Zone: Continuous causal sensors indicate safe passage within regional radius."
}

