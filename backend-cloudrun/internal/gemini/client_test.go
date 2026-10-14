package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai.coexistshield/backend/internal/model"
)

func TestGeminiModelVersionRequirement(t *testing.T) {
	if GeminiModelVersion != "gemini-3.8-flash" {
		t.Fatalf("expected GeminiModelVersion to strictly be 'gemini-3.8-flash', got '%s'", GeminiModelVersion)
	}
}

func TestGeminiApiKeyRequiredWhenMockDisabled(t *testing.T) {
	client := NewClientWithConfig(Config{
		ApiKey:      "",
		MockEnabled: false,
	})
	ctx := context.Background()

	// 1. TriageFieldPhoto
	_, err := client.TriageFieldPhoto(ctx, model.ForensicTriageRequest{
		ImageBase64:  "mockBase64",
		CategoryHint: "PHYSICAL_TRACE",
	})
	if !errors.Is(err, ErrApiKeyMissing) {
		t.Fatalf("expected ErrApiKeyMissing for TriageFieldPhoto, got %v", err)
	}

	// 2. ExtractIncidentEntities
	_, err = client.ExtractIncidentEntities(ctx, "クマの目撃情報", "HOTLINE", "JAPAN")
	if !errors.Is(err, ErrApiKeyMissing) {
		t.Fatalf("expected ErrApiKeyMissing for ExtractIncidentEntities, got %v", err)
	}

	// 3. GenerateExplainableAlert
	_, _, err = client.GenerateExplainableAlert(ctx, model.CausalAlertExplanationRequest{
		Tier:          "TIER_3_DANGER",
		TargetSpecies: "BEAR",
	}, nil)
	if !errors.Is(err, ErrApiKeyMissing) {
		t.Fatalf("expected ErrApiKeyMissing for GenerateExplainableAlert, got %v", err)
	}
}

func TestMultimodalForensicTriageTracesAndCarcassMock(t *testing.T) {
	client := NewClientWithConfig(Config{
		MockEnabled: true,
	})
	ctx := context.Background()

	// 1. Physical Trace (Claw mark on cedar)
	traceResp, err := client.TriageFieldPhoto(ctx, model.ForensicTriageRequest{
		ImageBase64:  "data:image/jpeg;base64,mockTraceData",
		CategoryHint: "PHYSICAL_TRACE",
		Latitude:     39.702,
		Longitude:    141.152,
	})
	if err != nil {
		t.Fatalf("unexpected error in trace triage: %v", err)
	}
	if traceResp.DetectedSpecies != "Ursus thibetanus" {
		t.Errorf("expected Ursus thibetanus, got %s", traceResp.DetectedSpecies)
	}
	if traceResp.Attributes == nil || traceResp.Attributes.TraceType == nil || *traceResp.Attributes.TraceType != "CLAW_MARK_BARK_STRIPPING" {
		t.Errorf("expected CLAW_MARK_BARK_STRIPPING attribute, got %+v", traceResp.Attributes)
	}
	if traceResp.Confidence < 0.90 {
		t.Errorf("expected confidence >= 0.90, got %f", traceResp.Confidence)
	}
	if traceResp.Source != "mock" {
		t.Errorf("expected Source == 'mock', got '%s'", traceResp.Source)
	}

	// 2. Roadkill carcass triage
	roadkillResp, err := client.TriageFieldPhoto(ctx, model.ForensicTriageRequest{
		ImageBase64:  "data:image/jpeg;base64,mockRoadkillData",
		CategoryHint: "ROADKILL",
		Latitude:     38.256,
		Longitude:    140.321,
	})
	if err != nil {
		t.Fatalf("unexpected error in roadkill triage: %v", err)
	}
	if roadkillResp.Attributes == nil || roadkillResp.Attributes.DecayState == nil || *roadkillResp.Attributes.DecayState != "FRESH" {
		t.Errorf("expected FRESH decay state, got %+v", roadkillResp.Attributes)
	}
	if roadkillResp.Source != "mock" {
		t.Errorf("expected Source == 'mock', got '%s'", roadkillResp.Source)
	}

	// 3. Attractant fruit tree
	treeResp, err := client.TriageFieldPhoto(ctx, model.ForensicTriageRequest{
		ImageBase64:  "data:image/jpeg;base64,mockTreeData",
		CategoryHint: "ATTRACTANT_FRUIT_TREE",
		Latitude:     39.710,
		Longitude:    141.160,
	})
	if err != nil {
		t.Fatalf("unexpected error in tree triage: %v", err)
	}
	if treeResp.DetectedSpecies != "Diospyros kaki" {
		t.Errorf("expected Diospyros kaki, got %s", treeResp.DetectedSpecies)
	}

	// 4. Livestock Depredation
	depResp, err := client.TriageFieldPhoto(ctx, model.ForensicTriageRequest{
		ImageBase64:  "data:image/jpeg;base64,mockDepredationData",
		CategoryHint: "LIVESTOCK_DEPREDATION",
		Latitude:     0.825,
		Longitude:    101.530,
	})
	if err != nil {
		t.Fatalf("unexpected error in depredation triage: %v", err)
	}
	if depResp.DetectedSpecies != "Panthera tigris sumatrae" {
		t.Errorf("expected Panthera tigris sumatrae, got %s", depResp.DetectedSpecies)
	}
}

func TestExtractIncidentEntitiesMultilingualCausalNLPMock(t *testing.T) {
	client := NewClientWithConfig(Config{
		MockEnabled: true,
	})
	ctx := context.Background()

	// 1. Japanese Municipal Disaster Radio Broadcast
	jpText := "盛岡市防災無線：本日午前7時20分頃、盛岡市松ヶ枝町付近（松ヶ枝小学校通学路付近）にて成獣のクマ1頭の目撃情報がありました。柿の木の周辺を徘徊しているため警戒してください。"
	jpEntity, err := client.ExtractIncidentEntities(ctx, jpText, "MUNICIPAL_DISASTER_RADIO", "JAPAN")
	if err != nil {
		t.Fatalf("unexpected error in JP NLP extraction: %v", err)
	}
	if jpEntity.Species != "BEAR" {
		t.Errorf("expected species BEAR, got %s", jpEntity.Species)
	}
	if jpEntity.ScientificName != "Ursus thibetanus" {
		t.Errorf("expected Ursus thibetanus, got %s", jpEntity.ScientificName)
	}
	if jpEntity.UrgencyLevel != "HIGH" {
		t.Errorf("expected HIGH urgency due to school route mention, got %s", jpEntity.UrgencyLevel)
	}
	if jpEntity.Confidence < 0.90 {
		t.Errorf("expected confidence >= 0.90, got %f", jpEntity.Confidence)
	}
	if jpEntity.InferenceProvider != "mock" {
		t.Errorf("expected InferenceProvider == 'mock', got '%s'", jpEntity.InferenceProvider)
	}

	// 2. English Social Media / Citizen Field Post (Tiger)
	enText := "Urgent: fresh tiger pugmarks and territorial scratches spotted along Block 14 oil palm plantation near Siak river corridor! Keep livestock corralled tonight."
	enEntity, err := client.ExtractIncidentEntities(ctx, enText, "BLUESKY", "INDONESIA")
	if err != nil {
		t.Fatalf("unexpected error in EN NLP extraction: %v", err)
	}
	if enEntity.Species != "TIGER" {
		t.Errorf("expected species TIGER, got %s", enEntity.Species)
	}
	if enEntity.IncidentType != "PHYSICAL_TRACE" {
		t.Errorf("expected PHYSICAL_TRACE incident type, got %s", enEntity.IncidentType)
	}

	// 3. Thai Crop-Raid Broadcast (Elephant)
	thText := "แจ้งเตือนช้างป่า 3 ตัว ออกจากป่าเขาอ่างฤๅไน มุ่งหน้าไร่อ้อยบ้านหนองคล้า ใกล้พื้นที่เก็บเกี่ยว"
	thEntity, err := client.ExtractIncidentEntities(ctx, thText, "REGIONAL_BOARD", "THAILAND")
	if err != nil {
		t.Fatalf("unexpected error in TH NLP extraction: %v", err)
	}
	if thEntity.Species != "ELEPHANT" {
		t.Errorf("expected species ELEPHANT, got %s", thEntity.Species)
	}
	if thEntity.IndividualCount != 3 {
		t.Errorf("expected 3 individuals, got %d", thEntity.IndividualCount)
	}
	if thEntity.IncidentType != "CROP_RAID" {
		t.Errorf("expected CROP_RAID incident type, got %s", thEntity.IncidentType)
	}
}

func TestGenerateExplainableAlertGemini38FlashMock(t *testing.T) {
	client := NewClientWithConfig(Config{
		MockEnabled: true,
	})
	ctx := context.Background()

	req := model.CausalAlertExplanationRequest{
		Tier:          "TIER_3_DANGER",
		TargetSpecies: "TIGER",
		Weather: map[string]interface{}{
			"windSpeedKmh":     14.0,
			"moonIllumination": 0.05,
		},
		NearbyAttractants: []map[string]interface{}{
			{"type": "KILL_SITE", "distanceMeters": 350},
		},
	}

	explanation, recs, err := client.GenerateExplainableAlert(ctx, req, nil)
	if err != nil {
		t.Fatalf("unexpected error in explanation: %v", err)
	}
	if explanation == "" {
		t.Errorf("expected non-empty explanation")
	}
	if len(recs) == 0 {
		t.Errorf("expected non-empty dispatch recommendations, got 0")
	}
	if recs[0].Priority != "IMMEDIATE" {
		t.Errorf("expected IMMEDIATE priority on primary recommendation, got %s", recs[0].Priority)
	}
}

func TestGeminiLiveHttpCallWithMockServer(t *testing.T) {
	sampleTriageJson := `{
		"detectedSpecies": "Ursus thibetanus",
		"commonName": "Asian Black Bear",
		"confidence": 0.98,
		"attributes": {
			"fruitLoadLevel": "ABUNDANT"
		},
		"suggestedTags": ["bear_sighting", "persimmon_lure"],
		"preliminaryThreatIndex": 0.91,
		"notes": "Real Gemini 3.8 Flash multimodal analysis result."
	}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "models/gemini-3.8-flash:generateContent") {
			t.Errorf("unexpected URL path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("key") != "secret-api-key" {
			t.Errorf("expected query param key=secret-api-key, got %s", r.URL.Query().Get("key"))
		}

		resp := geminiResponse{
			Candidates: []geminiCandidate{
				{
					Content: struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					}{
						Parts: []struct {
							Text string `json:"text"`
						}{
							{Text: sampleTriageJson},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClientWithConfig(Config{
		ApiKey:      "secret-api-key",
		Endpoint:    ts.URL,
		Model:       "gemini-3.8-flash",
		MockEnabled: false,
		Timeout:     5 * time.Second,
	})

	ctx := context.Background()
	res, err := client.TriageFieldPhoto(ctx, model.ForensicTriageRequest{
		ImageBase64:  "dummyBase64",
		CategoryHint: "ATTRACTANT_FRUIT_TREE",
		Latitude:     39.7,
		Longitude:    141.1,
	})
	if err != nil {
		t.Fatalf("expected successful live call, got error: %v", err)
	}
	if res.DetectedSpecies != "Ursus thibetanus" {
		t.Errorf("expected Ursus thibetanus, got %s", res.DetectedSpecies)
	}
	if res.Source != "gemini" {
		t.Errorf("expected Source == 'gemini', got '%s'", res.Source)
	}
}

func TestGeminiHttpErrorHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"code":403,"message":"API_KEY_INVALID"}}`, http.StatusForbidden)
	}))
	defer ts.Close()

	// 1. Without Mock -> Returns explicit error, never fakes success
	clientNoMock := NewClientWithConfig(Config{
		ApiKey:      "bad-key",
		Endpoint:    ts.URL,
		MockEnabled: false,
	})
	_, err := clientNoMock.TriageFieldPhoto(context.Background(), model.ForensicTriageRequest{
		ImageBase64: "dummy",
	})
	if err == nil {
		t.Fatalf("expected error when mock is disabled, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected error to contain 403 status, got %v", err)
	}

	// 2. With Mock -> Fallback explicitly marked as 'mock_fallback'
	clientWithMock := NewClientWithConfig(Config{
		ApiKey:      "bad-key",
		Endpoint:    ts.URL,
		MockEnabled: true,
	})
	res, err := clientWithMock.TriageFieldPhoto(context.Background(), model.ForensicTriageRequest{
		ImageBase64:  "dummy",
		CategoryHint: "ROADKILL",
	})
	if err != nil {
		t.Fatalf("expected fallback success when mock is enabled, got %v", err)
	}
	if res.Source != "mock_fallback" {
		t.Errorf("expected Source == 'mock_fallback', got '%s'", res.Source)
	}
}
