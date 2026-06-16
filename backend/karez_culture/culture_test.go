package karezculture

import (
	"karez-system/models"
	"testing"
)

func TestNewCultureService(t *testing.T) {
	service := New()
	if service == nil {
		t.Fatal("New() should return a non-nil CultureService")
	}
}

func TestGetTechnologyEvolution_Normal(t *testing.T) {
	service := New()
	analysis := service.GetTechnologyEvolution()

	if analysis == nil {
		t.Fatal("GetTechnologyEvolution() should return non-nil result")
	}

	if len(analysis.Evolutions) == 0 {
		t.Error("Evolutions should not be empty")
	}

	if len(analysis.KeyInnovations) == 0 {
		t.Error("KeyInnovations should not be empty")
	}

	if analysis.Summary == "" {
		t.Error("Summary should not be empty")
	}

	expectedEvolutions := 7
	if len(analysis.Evolutions) != expectedEvolutions {
		t.Errorf("Expected %d evolutions, got %d", expectedEvolutions, len(analysis.Evolutions))
	}

	expectedInnovations := 8
	if len(analysis.KeyInnovations) != expectedInnovations {
		t.Errorf("Expected %d innovations, got %d", expectedInnovations, len(analysis.KeyInnovations))
	}

	for i, era := range analysis.Evolutions {
		if era.Era == "" {
			t.Errorf("Evolution %d: Era should not be empty", i)
		}
		if era.TimePeriod == "" {
			t.Errorf("Evolution %d: TimePeriod should not be empty", i)
		}
		if len(era.KeyFeatures) == 0 {
			t.Errorf("Evolution %d: KeyFeatures should not be empty", i)
		}
		if era.AverageDepth <= 0 {
			t.Errorf("Evolution %d: AverageDepth should be positive, got %f", i, era.AverageDepth)
		}
		if era.AverageLength <= 0 {
			t.Errorf("Evolution %d: AverageLength should be positive, got %f", i, era.AverageLength)
		}
		if era.MaxFlowRate <= 0 {
			t.Errorf("Evolution %d: MaxFlowRate should be positive, got %f", i, era.MaxFlowRate)
		}
		if era.WaterLossRate < 0 || era.WaterLossRate > 100 {
			t.Errorf("Evolution %d: WaterLossRate should be between 0 and 100, got %f", i, era.WaterLossRate)
		}
	}

	for i, innovation := range analysis.KeyInnovations {
		if innovation.Name == "" {
			t.Errorf("Innovation %d: Name should not be empty", i)
		}
		if innovation.Impact < 0 || innovation.Impact > 10 {
			t.Errorf("Innovation %d: Impact should be between 0 and 10, got %f", i, innovation.Impact)
		}
	}
}

func TestGetTechnologyEvolution_ProgressValidation(t *testing.T) {
	service := New()
	analysis := service.GetTechnologyEvolution()

	eras := analysis.Evolutions

	for i := 1; i < len(eras); i++ {
		prev := eras[i-1]
		curr := eras[i]

		if curr.AverageDepth <= prev.AverageDepth {
			t.Errorf("Technical regression: AverageDepth should increase over time. Era[%d]=%f, Era[%d]=%f",
				i-1, prev.AverageDepth, i, curr.AverageDepth)
		}

		if curr.AverageLength < prev.AverageLength {
			t.Errorf("Technical regression: AverageLength should not decrease. Era[%d]=%f, Era[%d]=%f",
				i-1, prev.AverageLength, i, curr.AverageLength)
		}

		if curr.MaxFlowRate < prev.MaxFlowRate {
			t.Errorf("Technical regression: MaxFlowRate should not decrease. Era[%d]=%f, Era[%d]=%f",
				i-1, prev.MaxFlowRate, i, curr.MaxFlowRate)
		}

		if curr.WaterLossRate > prev.WaterLossRate {
			t.Errorf("Technical regression: WaterLossRate should not increase. Era[%d]=%f, Era[%d]=%f",
				i-1, prev.WaterLossRate, i, curr.WaterLossRate)
		}
	}

	first := eras[0]
	last := eras[len(eras)-1]

	depthRatio := last.AverageDepth / first.AverageDepth
	if depthRatio < 5 {
		t.Errorf("Technical progress insufficient: Depth ratio should be > 5, got %f", depthRatio)
	}

	lengthRatio := last.AverageLength / first.AverageLength
	if lengthRatio < 10 {
		t.Errorf("Technical progress insufficient: Length ratio should be > 10, got %f", lengthRatio)
	}

	lossReduction := first.WaterLossRate - last.WaterLossRate
	if lossReduction < 30 {
		t.Errorf("Water loss reduction insufficient: Should reduce by > 30%%, got %f%%", lossReduction)
	}
}

func TestGetCrossEraComparison_Normal(t *testing.T) {
	service := New()
	comparison := service.GetCrossEraComparison()

	if comparison == nil {
		t.Fatal("GetCrossEraComparison() should return non-nil result")
	}

	if comparison.KarezSystem.Name == "" {
		t.Error("KarezSystem name should not be empty")
	}

	if comparison.DripIrrigation.Name == "" {
		t.Error("DripIrrigation name should not be empty")
	}

	if len(comparison.ComparisonMetrics) == 0 {
		t.Error("ComparisonMetrics should not be empty")
	}

	if comparison.Conclusion == "" {
		t.Error("Conclusion should not be empty")
	}

	expectedMetrics := 10
	if len(comparison.ComparisonMetrics) != expectedMetrics {
		t.Errorf("Expected %d comparison metrics, got %d", expectedMetrics, len(comparison.ComparisonMetrics))
	}
}

func TestGetCrossEraComparison_WaterEfficiencyValidation(t *testing.T) {
	service := New()
	comparison := service.GetCrossEraComparison()

	karezEff := comparison.KarezSystem.WaterUseEfficiency
	dripEff := comparison.DripIrrigation.WaterUseEfficiency

	if karezEff < 70 || karezEff > 95 {
		t.Errorf("Karez water use efficiency should be realistic (70-95%%), got %f", karezEff)
	}

	if dripEff < 85 || dripEff > 99 {
		t.Errorf("Drip irrigation efficiency should be realistic (85-99%%), got %f", dripEff)
	}

	if dripEff <= karezEff {
		t.Error("Drip irrigation should have higher water use efficiency than karez")
	}

	var waterEffItem *models.ComparisonItem
	for i := range comparison.ComparisonMetrics {
		if comparison.ComparisonMetrics[i].Metric == "水资源利用效率" {
			waterEffItem = &comparison.ComparisonMetrics[i]
			break
		}
	}

	if waterEffItem == nil {
		t.Fatal("Could not find water efficiency comparison item")
	}

	if waterEffItem.BetterSolution != "滴灌" {
		t.Errorf("Expected drip to be better for water efficiency, got %s", waterEffItem.BetterSolution)
	}

	if waterEffItem.DripValue != dripEff {
		t.Errorf("Drip value mismatch: comparison=%f, metric=%f", waterEffItem.DripValue, dripEff)
	}

	if waterEffItem.KarezValue != karezEff {
		t.Errorf("Karez value mismatch: comparison=%f, metric=%f", waterEffItem.KarezValue, karezEff)
	}
}

func TestGetCrossEraComparison_EnergyConsumptionValidation(t *testing.T) {
	service := New()
	comparison := service.GetCrossEraComparison()

	karezEnergy := comparison.KarezSystem.EnergyConsumption
	dripEnergy := comparison.DripIrrigation.EnergyConsumption

	if karezEnergy < 0 || karezEnergy > 2 {
		t.Errorf("Karez energy consumption should be low (0-2 kWh/ha/day), got %f", karezEnergy)
	}

	if dripEnergy < 2 || dripEnergy > 20 {
		t.Errorf("Drip energy consumption should be higher (2-20 kWh/ha/day), got %f", dripEnergy)
	}

	if dripEnergy <= karezEnergy*5 {
		t.Error("Drip irrigation should consume significantly more energy than karez")
	}

	var energyItem *models.ComparisonItem
	for i := range comparison.ComparisonMetrics {
		if comparison.ComparisonMetrics[i].Metric == "能源消耗" {
			energyItem = &comparison.ComparisonMetrics[i]
			break
		}
	}

	if energyItem == nil {
		t.Fatal("Could not find energy consumption comparison item")
	}

	if energyItem.BetterSolution != "坎儿井" {
		t.Errorf("Expected karez to be better for energy consumption, got %s", energyItem.BetterSolution)
	}
}

func TestGetCrossEraComparison_BoundaryValues(t *testing.T) {
	service := New()
	comparison := service.GetCrossEraComparison()

	if comparison.KarezSystem.LifespanYears <= 0 {
		t.Error("Karez lifespan should be positive")
	}

	if comparison.DripIrrigation.LifespanYears <= 0 {
		t.Error("Drip lifespan should be positive")
	}

	if comparison.KarezSystem.LifespanYears <= comparison.DripIrrigation.LifespanYears {
		t.Error("Karez should have longer lifespan than drip irrigation")
	}

	karez := comparison.KarezSystem
	drip := comparison.DripIrrigation

	if karez.SetupCostPerHa <= 0 || drip.SetupCostPerHa <= 0 {
		t.Error("Setup costs should be positive")
	}

	if karez.MaintenanceCost <= 0 || drip.MaintenanceCost <= 0 {
		t.Error("Maintenance costs should be positive")
	}

	if drip.MaintenanceCost <= karez.MaintenanceCost {
		t.Error("Drip should have higher maintenance cost than karez")
	}

	if karez.CropYieldBoost < 0 || drip.CropYieldBoost < 0 {
		t.Error("Crop yield boosts should not be negative")
	}

	if drip.CropYieldBoost <= karez.CropYieldBoost {
		t.Error("Drip should have higher crop yield boost than karez")
	}
}

func TestGetCrossEraComparison_Consistency(t *testing.T) {
	service := New()
	comparison := service.GetCrossEraComparison()

	metricMap := make(map[string]bool)
	for _, item := range comparison.ComparisonMetrics {
		if item.Metric == "" {
			t.Error("Metric name should not be empty")
		}
		if metricMap[item.Metric] {
			t.Errorf("Duplicate metric: %s", item.Metric)
		}
		metricMap[item.Metric] = true

		if item.BetterSolution != "坎儿井" && item.BetterSolution != "滴灌" {
			t.Errorf("Invalid BetterSolution for %s: %s", item.Metric, item.BetterSolution)
		}

		if item.Notes == "" {
			t.Errorf("Notes should not be empty for metric: %s", item.Metric)
		}
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
			t.Errorf("Missing expected metric: %s", expected)
		}
	}

	karezWins := 0
	dripWins := 0
	for _, item := range comparison.ComparisonMetrics {
		if item.BetterSolution == "坎儿井" {
			karezWins++
		} else {
			dripWins++
		}
	}

	if karezWins < 4 {
		t.Errorf("Karez should win at least 4 metrics, won %d", karezWins)
	}
	if dripWins < 3 {
		t.Errorf("Drip should win at least 3 metrics, won %d", dripWins)
	}
}

func BenchmarkGetTechnologyEvolution(b *testing.B) {
	service := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetTechnologyEvolution()
	}
}

func BenchmarkGetCrossEraComparison(b *testing.B) {
	service := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetCrossEraComparison()
	}
}
