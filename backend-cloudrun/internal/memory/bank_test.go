package memory

import (
	"strings"
	"testing"

	"ai.coexistshield/backend/internal/model"
)

func TestBankInitializationAndEmbedding(t *testing.T) {
	bank := NewBank()
	bank.LoadFixtureMemory()
	if bank == nil {
		t.Fatalf("expected non-nil memory bank")
	}

	if len(bank.memories) == 0 {
		t.Fatalf("expected seeded historical incidents, got 0")
	}

	for _, m := range bank.memories {
		if len(m.EmbeddingVector) != 768 {
			t.Errorf("expected 768-dim embedding for memory %s, got %d", m.MemoryId, len(m.EmbeddingVector))
		}
	}
}

func TestSearchAnalogousIncidentsBearAttack(t *testing.T) {
	bank := NewBank()
	bank.LoadFixtureMemory()

	req := model.MemorySearchRequest{
		QueryText:     "Asian Black Bear foraging on abandoned persimmon fruit near school route during new moon",
		Region:        "JAPAN",
		Species:       "BEAR",
		TopK:          2,
		MinSimilarity: 0.50,
	}

	resp := bank.SearchAnalogousIncidents(req)

	if !resp.QueryVectorGenerated {
		t.Errorf("expected QueryVectorGenerated to be true")
	}
	if len(resp.Matches) == 0 {
		t.Fatalf("expected matches for bear search, got 0")
	}

	topMatch := resp.Matches[0]
	if topMatch.Species != "BEAR" {
		t.Errorf("expected top match species BEAR, got %s", topMatch.Species)
	}
	if topMatch.CosineSimilarity <= 0.50 {
		t.Errorf("expected similarity > 0.50, got %f", topMatch.CosineSimilarity)
	}
	if resp.ElevatedTier == "" {
		t.Errorf("expected autonomous risk elevation tier from severe analog")
	}
}

func TestSearchAnalogousIncidentsTigerAmbush(t *testing.T) {
	bank := NewBank()
	bank.LoadFixtureMemory()

	req := model.MemorySearchRequest{
		QueryText:     "Sumatran tiger returning to cached domestic cow carcass in pasture scrub buffer",
		Region:        "INDONESIA",
		Species:       "TIGER",
		TopK:          1,
		MinSimilarity: 0.50,
	}

	resp := bank.SearchAnalogousIncidents(req)
	if len(resp.Matches) == 0 {
		t.Fatalf("expected tiger match, got 0")
	}

	topMatch := resp.Matches[0]
	if topMatch.MemoryId != "mem_id_01" {
		t.Errorf("expected mem_id_01 for tiger kill-site analog, got %s", topMatch.MemoryId)
	}
	if resp.ElevatedTier != "TIER_3_DANGER" {
		t.Errorf("expected TIER_3_DANGER elevation, got %s", resp.ElevatedTier)
	}
	if !strings.Contains(resp.ElevationReason, "Tier 3 Danger") {
		t.Errorf("expected elevation reason mentioning Tier 3 Danger, got %s", resp.ElevationReason)
	}
}

func TestSearchAnalogousIncidentsElephantCropRaid(t *testing.T) {
	bank := NewBank()
	bank.LoadFixtureMemory()

	req := model.MemorySearchRequest{
		QueryText:     "Elephant herd approaching ripening sugarcane plantation border during nighttime",
		Region:        "THAILAND",
		Species:       "ELEPHANT",
		TopK:          1,
		MinSimilarity: 0.50,
	}

	resp := bank.SearchAnalogousIncidents(req)
	if len(resp.Matches) == 0 {
		t.Fatalf("expected elephant match, got 0")
	}

	topMatch := resp.Matches[0]
	if topMatch.MemoryId != "mem_th_01" {
		t.Errorf("expected mem_th_01, got %s", topMatch.MemoryId)
	}
	if topMatch.Category != "CROP_RAID" {
		t.Errorf("expected category CROP_RAID, got %s", topMatch.Category)
	}
}
