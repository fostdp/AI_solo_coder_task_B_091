package eracomparator

import (
	"testing"
)

func TestNewEraComparator(t *testing.T) {
	comparator := New()
	if comparator == nil {
		t.Fatal("Expected non-nil EraComparator")
	}
}

func TestCompare_Normal(t *testing.T) {
	comparator := New()
	result := comparator.Compare()

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.KarezSystem.Name == "" {
		t.Error("Karez system name should not be empty")
	}

	if result.DripIrrigation.Name == "" {
		t.Error("Drip irrigation name should not be empty")
	}

	if len(result.ComparisonMetrics) == 0 {
		t.Error("Expected non-empty comparison metrics")
	}

	if result.Conclusion == "" {
		t.Error("Conclusion should not be empty")
	}
}

func TestCompare_WaterEfficiencyValidation(t *testing.T) {
	comparator := New()
	comparison := comparator.Compare()

	karezEff := comparison.KarezSystem.WaterUseEfficiency
	dripEff := comparison.DripIrrigation.WaterUseEfficiency

	if karezEff < 30 || karezEff > 60 {
		t.Errorf("Karez water use efficiency should be realistic (30-60%%), got %f", karezEff)
	}

	if dripEff < 85 || dripEff > 99 {
		t.Errorf("Drip irrigation efficiency should be realistic (85-99%%), got %f", dripEff)
	}

	if dripEff <= karezEff {
		t.Error("Drip irrigation should have higher water use efficiency than karez")
	}
}

func TestCompare_EnergyConsumptionValidation(t *testing.T) {
	comparator := New()
	comparison := comparator.Compare()

	karezEnergy := comparison.KarezSystem.EnergyConsumption
	dripEnergy := comparison.DripIrrigation.EnergyConsumption

	if karezEnergy != 0 {
		t.Errorf("Karez energy consumption should be 0 (gravity flow), got %f", karezEnergy)
	}

	if dripEnergy <= 0 {
		t.Errorf("Drip irrigation should consume energy, got %f", dripEnergy)
	}

	if dripEnergy < 5 || dripEnergy > 15 {
		t.Errorf("Drip irrigation energy should be in realistic range (5-15 kWh/ha/day), got %f", dripEnergy)
	}
}

func TestCompare_BoundaryValues(t *testing.T) {
	comparator := New()
	comparison := comparator.Compare()

	for _, item := range comparison.ComparisonMetrics {
		if item.Metric == "" {
			t.Error("Metric name should not be empty")
		}

		if item.BetterSolution != "坎儿井" && item.BetterSolution != "滴灌" {
			t.Errorf("Invalid BetterSolution for %s: %s", item.Metric, item.BetterSolution)
		}

		if item.Notes == "" {
			t.Errorf("Notes should not be empty for metric: %s", item.Metric)
		}
	}
}

func TestCompare_Consistency(t *testing.T) {
	comparator := New()
	comparison := comparator.Compare()

	metricMap := make(map[string]bool)
	for _, item := range comparison.ComparisonMetrics {
		if metricMap[item.Metric] {
			t.Errorf("Duplicate metric: %s", item.Metric)
		}
		metricMap[item.Metric] = true
	}

	expectedMetrics := []string{
		"水资源利用效率",
		"能源消耗",
		"初始建设成本",
		"年维护成本",
		"作物增产幅度",
		"使用寿命",
		"生态环境影响",
		"气候适应性",
		"文化遗产价值",
		"综合经济回报（50年周期）",
	}

	for _, expected := range expectedMetrics {
		if !metricMap[expected] {
			t.Errorf("Expected metric not found: %s", expected)
		}
	}
}

func TestCompare_StandardBasis(t *testing.T) {
	comparator := New()
	comparison := comparator.Compare()

	if comparison.DripIrrigation.StandardBasis == "" {
		t.Error("Drip irrigation should have standard basis")
	}

	if comparison.DripIrrigation.ApplicableConditions == "" {
		t.Error("Drip irrigation should have applicable conditions")
	}

	if len(comparison.ComparisonStandards) == 0 {
		t.Error("Expected non-empty comparison standards")
	}

	if comparison.ScopeOfApplication == "" {
		t.Error("Expected non-empty scope of application")
	}
}

func TestCompare_LifespanValidation(t *testing.T) {
	comparator := New()
	comparison := comparator.Compare()

	karezLifespan := comparison.KarezSystem.LifespanYears
	dripLifespan := comparison.DripIrrigation.LifespanYears

	if karezLifespan <= dripLifespan {
		t.Errorf("Karez should have longer lifespan than drip irrigation: karez=%d, drip=%d",
			karezLifespan, dripLifespan)
	}

	if karezLifespan < 50 || karezLifespan > 200 {
		t.Errorf("Karez lifespan should be realistic (50-200 years), got %d", karezLifespan)
	}

	if dripLifespan < 5 || dripLifespan > 30 {
		t.Errorf("Drip lifespan should be realistic (5-30 years), got %d", dripLifespan)
	}
}

func TestCompare_CostValidation(t *testing.T) {
	comparator := New()
	comparison := comparator.Compare()

	if comparison.KarezSystem.SetupCostPerHa <= 0 {
		t.Error("Karez setup cost should be positive")
	}

	if comparison.DripIrrigation.SetupCostPerHa <= 0 {
		t.Error("Drip setup cost should be positive")
	}
}

func TestCompare_KarezMetricsStructure(t *testing.T) {
	comparator := New()
	comparison := comparator.Compare()

	karez := comparison.KarezSystem

	if karez.TechnologyLevel == "" {
		t.Error("Technology level should not be empty")
	}

	if karez.EcosystemImpact == "" {
		t.Error("Ecosystem impact should not be empty")
	}
}

func TestCompare_DripMetricsStructure(t *testing.T) {
	comparator := New()
	comparison := comparator.Compare()

	drip := comparison.DripIrrigation

	if drip.Description == "" {
		t.Error("Description should not be empty")
	}

	if drip.SystemType == "" {
		t.Error("System type should not be empty")
	}
}
