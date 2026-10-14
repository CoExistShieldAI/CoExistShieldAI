package gemini

import (
	"context"
	"testing"

	"ai.coexistshield/backend/internal/model"
)

func TestGeminiModelVersionRequirement(t *testing.T) {
	if GeminiModelVersion != "gemini-3.8-flash" {
		t.Fatalf("expected GeminiModelVersion to strictly be 'gemini-3.8-flash', got '%s'", GeminiModelVersion)
	}
}

func TestMultimodalForensicTriageTracesAndCarcass(t *testing.T) {
	client := NewClient()
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

func TestExtractIncidentEntitiesMultilingualCausalNLP(t *testing.T) {
	client := NewClient()
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

func TestGenerateExplainableAlertGemini38Flash(t *testing.T) {
	client := NewClient()
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
