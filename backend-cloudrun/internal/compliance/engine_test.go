package compliance

import (
	"strings"
	"testing"
)

func TestRspoAuditPackageAndCsv(t *testing.T) {
	eng := NewEngine()

	pkg, err := eng.GetRspoAuditPackage("wilmar_sumatra_02")
	if err != nil {
		t.Fatalf("failed to get rspo audit package: %v", err)
	}

	if pkg.PlantationID != "wilmar_sumatra_02" {
		t.Errorf("expected wilmar_sumatra_02, got %s", pkg.PlantationID)
	}
	if pkg.HumanFatalities != 0 || pkg.WildlifeFatalities != 0 {
		t.Errorf("expected 0 fatalities for Zero-Harm certification, got %d human, %d wildlife",
			pkg.HumanFatalities, pkg.WildlifeFatalities)
	}
	if pkg.VerifiedSafePassages < 100 {
		t.Errorf("expected >= 100 verified safe passages, got %d", pkg.VerifiedSafePassages)
	}
	if pkg.ZeroHarmComplianceScore != 1.0 {
		t.Errorf("expected 1.0 compliance score, got %.2f", pkg.ZeroHarmComplianceScore)
	}
	if len(pkg.Sha256AuditSignature) != 64 {
		t.Errorf("expected 64-char sha256 signature, got %s", pkg.Sha256AuditSignature)
	}

	csv, err := eng.GetRspoAuditCsv("wilmar_sumatra_02")
	if err != nil {
		t.Fatalf("failed to generate CSV: %v", err)
	}
	if !strings.Contains(csv, "wilmar_sumatra_02") {
		t.Errorf("expected plantation ID in CSV: %s", csv)
	}
	if !strings.Contains(csv, "CERTIFIED_ZERO_HARM_HCV") {
		t.Errorf("expected status in CSV: %s", csv)
	}
}

func TestGcfBudgetImpact(t *testing.T) {
	eng := NewEngine()

	resp, err := eng.GetGcfBudgetImpact("03201")
	if err != nil {
		t.Fatalf("failed to get gcf budget impact: %v", err)
	}

	if resp.MunicipalityCode != "03201" {
		t.Errorf("expected 03201, got %s", resp.MunicipalityCode)
	}
	if resp.CurrentRaisedYen < 10000000 {
		t.Errorf("expected >= 10M yen raised, got %d", resp.CurrentRaisedYen)
	}
	if resp.CostPerTreeYen != 35000 {
		t.Errorf("expected 35000 yen cost per tree, got %d", resp.CostPerTreeYen)
	}
	if resp.CompletedFelledTreesCount != 214 {
		t.Errorf("expected 214 initially completed trees, got %d", resp.CompletedFelledTreesCount)
	}
	if resp.QuantifiedBearEncounterRiskReduction < 40.0 {
		t.Errorf("expected >= 40%% encounter risk reduction, got %.1f", resp.QuantifiedBearEncounterRiskReduction)
	}
	if resp.SchoolChildrenProtectedCount != 3420 {
		t.Errorf("expected 3420 protected school children, got %d", resp.SchoolChildrenProtectedCount)
	}
	if len(resp.RecentFellingLogs) == 0 {
		t.Errorf("expected at least one recent felling log")
	}
}

func TestPriorityKakiLayer(t *testing.T) {
	eng := NewEngine()

	layer, err := eng.GetPriorityKakiLayer("03201")
	if err != nil {
		t.Fatalf("failed to get priority kaki layer: %v", err)
	}

	if len(layer.PriorityTrees) < 3 {
		t.Fatalf("expected at least 3 priority trees, got %d", len(layer.PriorityTrees))
	}

	first := layer.PriorityTrees[0]
	if first.TreeID != "kt_1120" {
		t.Errorf("expected kt_1120 first, got %s", first.TreeID)
	}
	if first.HazardScore < 0.90 {
		t.Errorf("expected hazard score >= 0.90, got %.2f", first.HazardScore)
	}
	if first.DistanceToSchoolRouteMeters != 120 {
		t.Errorf("expected 120m school distance, got %d", first.DistanceToSchoolRouteMeters)
	}
	if first.OwnerConsentStatus != "GRANTED" {
		t.Errorf("expected GRANTED owner consent status, got %s", first.OwnerConsentStatus)
	}
}

func TestMunicipalActionRBAC(t *testing.T) {
	eng := NewEngine()

	// 1. Citizen cannot log felling (tampering prevention)
	citizenReq := MunicipalKakiActionRequest{
		TreeID:       "kt_1120",
		Action:       "FELLED",
		OfficerID:    "citizen_user_99",
		OfficerRole:  "CITIZEN",
		PermitNumber: "UNAUTHORIZED",
	}
	_, err := eng.LogMunicipalAction(citizenReq)
	if err == nil {
		t.Fatalf("expected RBAC error for CITIZEN role, got nil")
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("expected forbidden message, got %v", err)
	}

	// 2. Official logs valid felling action
	officialReq := MunicipalKakiActionRequest{
		TreeID:            "kt_1120",
		Action:            "FELLED",
		OfficerID:         "OFFICER-IWATE-01",
		OfficerRole:       "OFFICIAL",
		PermitNumber:      "PERMIT-MRK-2026-1120",
		BiomassDisposedKg: 520.0,
		Notes:             "Complete tree felling and biomass chipped on-site.",
	}
	resp, err := eng.LogMunicipalAction(officialReq)
	if err != nil {
		t.Fatalf("expected success for OFFICIAL role, got: %v", err)
	}

	if resp.UpdatedTree.ManagementStatus != "FELLED" {
		t.Errorf("expected FELLED status, got %s", resp.UpdatedTree.ManagementStatus)
	}
	if resp.UpdatedTree.HazardScore != 0.0 {
		t.Errorf("expected 0.0 hazard score after felling, got %.2f", resp.UpdatedTree.HazardScore)
	}
	if resp.TreeRiskDropPct != 100.0 {
		t.Errorf("expected 100%% risk drop, got %.1f", resp.TreeRiskDropPct)
	}
	if resp.GcfCompletedFelledTrees != 215 {
		t.Errorf("expected 215 completed felled trees, got %d", resp.GcfCompletedFelledTrees)
	}
	if resp.NewMunicipalityRiskScore >= 0.90 {
		t.Errorf("expected downward recalibrated risk score, got %.2f", resp.NewMunicipalityRiskScore)
	}
}
