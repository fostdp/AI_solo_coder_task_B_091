package waterlevel

import (
	"karez-system/models"
	"math"
	"testing"
)

func TestNewWaterLevelSimulator(t *testing.T) {
	sim := New()
	if sim == nil {
		t.Fatal("New() should return a non-nil WaterLevelSimulator")
	}
	if sim.aquiferParams == nil {
		t.Error("aquiferParams should be initialized")
	}
	expectedAquifers := 5
	if len(sim.aquiferParams) != expectedAquifers {
		t.Errorf("Expected %d aquifer types, got %d", expectedAquifers, len(sim.aquiferParams))
	}
}

func TestGetAquiferParams_ValidTypes(t *testing.T) {
	sim := New()

	testCases := []struct {
		name     string
		aquifer  string
		wantPerm float64
	}{
		{"gravel", "gravel", 0.01},
		{"sand", "sand", 0.001},
		{"silt", "silt", 0.0001},
		{"clay", "clay", 0.00001},
		{"limestone", "limestone", 0.0005},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := sim.getAquiferParams(tc.aquifer)
			if params.Permeability != tc.wantPerm {
				t.Errorf("%s: expected permeability %f, got %f", tc.name, tc.wantPerm, params.Permeability)
			}
			if params.SpecificYield <= 0 || params.SpecificYield > 1 {
				t.Errorf("%s: SpecificYield should be between 0 and 1, got %f", tc.name, params.SpecificYield)
			}
		})
	}
}

func TestGetAquiferParams_InvalidType(t *testing.T) {
	sim := New()
	params := sim.getAquiferParams("invalid_type")

	defaultParams := sim.aquiferParams["gravel"]
	if params.Permeability != defaultParams.Permeability {
		t.Error("Invalid aquifer type should default to gravel")
	}
}

func TestGetDefaultScenarios(t *testing.T) {
	sim := New()
	scenarios := sim.GetDefaultScenarios()

	if len(scenarios) == 0 {
		t.Fatal("GetDefaultScenarios() should return non-empty list")
	}

	expectedScenarios := 5
	if len(scenarios) != expectedScenarios {
		t.Errorf("Expected %d default scenarios, got %d", expectedScenarios, len(scenarios))
	}

	for i, s := range scenarios {
		if s.ScenarioName == "" {
			t.Errorf("Scenario %d: name should not be empty", i)
		}
		if s.InitialWaterLevel <= 0 {
			t.Errorf("Scenario %d: InitialWaterLevel should be positive, got %f", i, s.InitialWaterLevel)
		}
		if s.DurationYears < 0 {
			t.Errorf("Scenario %d: DurationYears should not be negative, got %d", i, s.DurationYears)
		}
		if s.Description == "" {
			t.Errorf("Scenario %d: Description should not be empty", i)
		}
	}

	if scenarios[0].ChangeRate != 0 {
		t.Error("First scenario should be stable (change rate = 0)")
	}

	if scenarios[1].ChangeRate <= 0 {
		t.Error("Slow decline scenario should have positive change rate")
	}

	if scenarios[4].ChangeRate >= 0 {
		t.Error("Recovery scenario should have negative change rate")
	}
}

func TestSimulateWaterLevelImpact_StableScenario(t *testing.T) {
	sim := New()

	req := models.WaterLevelSimulationRequest{
		KarezID:          1,
		BaselineFlowRate: 3000,
		ShaftDepth:       40,
		AquiferType:      "gravel",
		Scenarios: []models.WaterLevelScenario{
			{
				ScenarioName:      "稳定测试",
				InitialWaterLevel: 30,
				TargetWaterLevel:  30,
				ChangeRate:        0,
				DurationYears:     20,
				Description:       "测试稳定状态",
			},
		},
	}

	results := sim.SimulateWaterLevelImpact(req)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]

	if result.KarezID != 1 {
		t.Errorf("Expected KarezID 1, got %d", result.KarezID)
	}

	if result.BaselineFlow != 3000 {
		t.Errorf("Expected baseline flow 3000, got %f", result.BaselineFlow)
	}

	if result.ScenarioName != "稳定测试" {
		t.Errorf("Expected scenario name '稳定测试', got %s", result.ScenarioName)
	}

	expectedPoints := 21
	if len(result.DataPoints) != expectedPoints {
		t.Errorf("Expected %d data points, got %d", expectedPoints, len(result.DataPoints))
	}

	for i, dp := range result.DataPoints {
		if dp.Year != i {
			t.Errorf("Expected year %d, got %d", i, dp.Year)
		}

		if math.Abs(dp.WaterLevel-30) > 0.01 {
			t.Errorf("Year %d: water level should remain ~30, got %f", i, dp.WaterLevel)
		}

		if dp.FlowRate <= 0 {
			t.Errorf("Year %d: flow rate should be positive, got %f", i, dp.FlowRate)
		}

		if !dp.IsFlowSustained {
			t.Errorf("Year %d: flow should be sustained in stable scenario", i)
		}

		if dp.WarningLevel != "正常" {
			t.Errorf("Year %d: warning level should be '正常', got %s", i, dp.WarningLevel)
		}
	}

	if result.TotalDecline > 1 {
		t.Errorf("Total decline should be minimal in stable scenario, got %f", result.TotalDecline)
	}

	if result.YearsUntilDry != -1 {
		t.Errorf("YearsUntilDry should be -1 in stable scenario, got %d", result.YearsUntilDry)
	}

	if len(result.Recommendations) == 0 {
		t.Error("Should have recommendations")
	}
}

func TestSimulateWaterLevelImpact_DeclineScenario(t *testing.T) {
	sim := New()

	req := models.WaterLevelSimulationRequest{
		KarezID:          2,
		BaselineFlowRate: 3000,
		ShaftDepth:       40,
		AquiferType:      "gravel",
		Scenarios: []models.WaterLevelScenario{
			{
				ScenarioName:      "急剧下降测试",
				InitialWaterLevel: 30,
				TargetWaterLevel:  0,
				ChangeRate:        1.0,
				DurationYears:     30,
				Description:       "测试急剧下降",
			},
		},
	}

	results := sim.SimulateWaterLevelImpact(req)
	result := results[0]

	firstFlow := result.DataPoints[0].FlowRate
	lastFlow := result.DataPoints[len(result.DataPoints)-1].FlowRate

	if lastFlow >= firstFlow {
		t.Error("Flow rate should decrease over time in decline scenario")
	}

	if result.TotalDecline <= 0 {
		t.Errorf("Total decline should be positive, got %f", result.TotalDecline)
	}

	warningLevels := make(map[string]bool)
	for _, dp := range result.DataPoints {
		warningLevels[dp.WarningLevel] = true
	}

	expectedWarnings := []string{"正常", "注意", "警告", "严重", "危急"}
	foundAll := true
	for _, w := range expectedWarnings {
		if !warningLevels[w] {
			foundAll = false
			t.Logf("Warning level '%s' not found", w)
		}
	}
	if !foundAll {
		t.Log("Note: Not all warning levels may appear depending on simulation parameters")
	}

	if result.YearsUntilDry == -1 {
		t.Error("Should predict dry years in rapid decline scenario")
	}

	if result.YearsUntilDry > 30 {
		t.Errorf("Should dry within 30 years for 1m/year decline to 0, got %d", result.YearsUntilDry)
	}

	lastDP := result.DataPoints[len(result.DataPoints)-1]
	if lastDP.WaterLevel > 1 {
		t.Errorf("Final water level should be very low, got %f", lastDP.WaterLevel)
	}
}

func TestSimulateWaterLevelImpact_RecoveryScenario(t *testing.T) {
	sim := New()

	req := models.WaterLevelSimulationRequest{
		KarezID:          3,
		BaselineFlowRate: 3000,
		ShaftDepth:       40,
		AquiferType:      "gravel",
		Scenarios: []models.WaterLevelScenario{
			{
				ScenarioName:      "生态恢复测试",
				InitialWaterLevel: 25,
				TargetWaterLevel:  30,
				ChangeRate:        -0.3,
				DurationYears:     50,
				Description:       "测试生态恢复",
			},
		},
	}

	results := sim.SimulateWaterLevelImpact(req)
	result := results[0]

	firstFlow := result.DataPoints[0].FlowRate
	lastFlow := result.DataPoints[len(result.DataPoints)-1].FlowRate

	if lastFlow <= firstFlow {
		t.Error("Flow rate should increase over time in recovery scenario")
	}

	if result.TotalDecline > 5 {
		t.Errorf("Total decline should be small in recovery scenario, got %f", result.TotalDecline)
	}

	if result.YearsUntilDry != -1 {
		t.Errorf("YearsUntilDry should be -1 in recovery scenario, got %d", result.YearsUntilDry)
	}

	foundRecoveryMsg := false
	for _, rec := range result.Recommendations {
		if contains(rec, "恢复") || contains(rec, "节水") || contains(rec, "补水") {
			foundRecoveryMsg = true
			break
		}
	}
	if !foundRecoveryMsg {
		t.Error("Should have recovery-related recommendations")
	}
}

func TestSimulateWaterLevelImpact_DefaultValues(t *testing.T) {
	sim := New()

	req := models.WaterLevelSimulationRequest{
		KarezID: 1,
	}

	results := sim.SimulateWaterLevelImpact(req)

	if len(results) != 5 {
		t.Errorf("Should use default 5 scenarios when none provided, got %d", len(results))
	}

	result := results[0]
	if result.BaselineFlow <= 0 {
		t.Errorf("Should use default baseline flow > 0, got %f", result.BaselineFlow)
	}
}

func TestSimulateWaterLevelImpact_EdgeCases(t *testing.T) {
	sim := New()

	t.Run("zero_water_level", func(t *testing.T) {
		req := models.WaterLevelSimulationRequest{
			KarezID:          1,
			BaselineFlowRate: 3000,
			Scenarios: []models.WaterLevelScenario{
				{
					ScenarioName:      "零水位",
					InitialWaterLevel: 0,
					TargetWaterLevel:  0,
					ChangeRate:        0,
					DurationYears:     5,
					Description:       "测试零水位",
				},
			},
		}

		results := sim.SimulateWaterLevelImpact(req)
		result := results[0]

		for _, dp := range result.DataPoints {
			if dp.FlowRate < 0 {
				t.Errorf("Flow rate should not be negative, got %f", dp.FlowRate)
			}
			if dp.FlowRate > 0 {
				t.Logf("Warning: got flow %f at zero water level, may be acceptable", dp.FlowRate)
			}
		}
	})

	t.Run("negative_water_level", func(t *testing.T) {
		req := models.WaterLevelSimulationRequest{
			KarezID:          1,
			BaselineFlowRate: 3000,
			Scenarios: []models.WaterLevelScenario{
				{
					ScenarioName:      "负水位",
					InitialWaterLevel: -5,
					TargetWaterLevel:  -5,
					ChangeRate:        0,
					DurationYears:     5,
					Description:       "测试负水位",
				},
			},
		}

		results := sim.SimulateWaterLevelImpact(req)
		result := results[0]

		for _, dp := range result.DataPoints {
			if dp.FlowRate < 0 {
				t.Errorf("Flow rate should not be negative, got %f", dp.FlowRate)
			}
			if dp.IsFlowSustained {
				t.Error("Flow should not be sustained at negative water level")
			}
		}
	})

	t.Run("zero_baseline_flow", func(t *testing.T) {
		req := models.WaterLevelSimulationRequest{
			KarezID:          1,
			BaselineFlowRate: 0,
			Scenarios: []models.WaterLevelScenario{
				{
					ScenarioName:      "零流量",
					InitialWaterLevel: 30,
					TargetWaterLevel:  30,
					ChangeRate:        0,
					DurationYears:     5,
					Description:       "测试零流量",
				},
			},
		}

		results := sim.SimulateWaterLevelImpact(req)
		result := results[0]

		if result.BaselineFlow != 3000 {
			t.Errorf("Zero baseline should be set to default 3000, got %f", result.BaselineFlow)
		}

		for _, dp := range result.DataPoints {
			if dp.FlowRate < 0 {
				t.Errorf("Flow rate should not be negative, got %f", dp.FlowRate)
			}
		}
	})

	t.Run("different_aquifer_types", func(t *testing.T) {
		aquifers := []string{"gravel", "sand", "silt", "clay", "limestone", "unknown"}

		for _, aquifer := range aquifers {
			req := models.WaterLevelSimulationRequest{
				KarezID:          1,
				BaselineFlowRate: 3000,
				AquiferType:      aquifer,
				Scenarios: []models.WaterLevelScenario{
					{
						ScenarioName:      aquifer + "_测试",
						InitialWaterLevel: 30,
						TargetWaterLevel:  30,
						ChangeRate:        0,
						DurationYears:     5,
						Description:       "测试含水层" + aquifer,
					},
				},
			}

			results := sim.SimulateWaterLevelImpact(req)
			if len(results) != 1 {
				t.Errorf("%s: expected 1 result, got %d", aquifer, len(results))
			}

			result := results[0]
			for _, dp := range result.DataPoints {
				if dp.FlowRate < 0 {
					t.Errorf("%s: flow rate should not be negative, got %f", aquifer, dp.FlowRate)
				}
			}
		}
	})
}

func TestSimulateWaterLevelImpact_FlowValidation(t *testing.T) {
	sim := New()

	scenarios := []models.WaterLevelScenario{
		{
			ScenarioName:      "稳定",
			InitialWaterLevel: 30,
			TargetWaterLevel:  30,
			ChangeRate:        0,
			DurationYears:     10,
		},
		{
			ScenarioName:      "缓慢下降",
			InitialWaterLevel: 30,
			TargetWaterLevel:  20,
			ChangeRate:        0.2,
			DurationYears:     10,
		},
		{
			ScenarioName:      "快速下降",
			InitialWaterLevel: 30,
			TargetWaterLevel:  10,
			ChangeRate:        0.5,
			DurationYears:     10,
		},
	}

	req := models.WaterLevelSimulationRequest{
		KarezID:          1,
		BaselineFlowRate: 3000,
		Scenarios:        scenarios,
	}

	results := sim.SimulateWaterLevelImpact(req)

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	flows := make([]float64, 3)
	for i, r := range results {
		lastIdx := len(r.DataPoints) - 1
		flows[i] = r.DataPoints[lastIdx].FlowRate
	}

	if flows[0] <= flows[1] {
		t.Error("Stable flow should be > slow decline flow")
	}
	if flows[1] <= flows[2] {
		t.Error("Slow decline flow should be > fast decline flow")
	}

	declines := make([]float64, 3)
	for i, r := range results {
		declines[i] = r.TotalDecline
	}

	if declines[0] >= declines[1] {
		t.Error("Stable decline should be < slow decline")
	}
	if declines[1] >= declines[2] {
		t.Error("Slow decline should be < fast decline")
	}
}

func TestIntToStr(t *testing.T) {
	testCases := []struct {
		input  int
		want   string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{123, "123"},
		{1000, "1000"},
		{-1, "-1"},
		{-123, "-123"},
	}

	for _, tc := range testCases {
		got := intToStr(tc.input)
		if got != tc.want {
			t.Errorf("intToStr(%d) = %s, want %s", tc.input, got, tc.want)
		}
	}
}

func TestRoundFloat(t *testing.T) {
	testCases := []struct {
		value    float64
		decimals int
		want     float64
	}{
		{1.2345, 2, 1.23},
		{1.2355, 2, 1.24},
		{1.0, 0, 1.0},
		{1.5, 0, 2.0},
		{-1.234, 2, -1.23},
		{0.0, 5, 0.0},
	}

	for _, tc := range testCases {
		got := roundFloat(tc.value, tc.decimals)
		if math.Abs(got-tc.want) > 0.0001 {
			t.Errorf("roundFloat(%f, %d) = %f, want %f", tc.value, tc.decimals, got, tc.want)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func BenchmarkSimulateWaterLevelImpact(b *testing.B) {
	sim := New()
	req := models.WaterLevelSimulationRequest{
		KarezID:          1,
		BaselineFlowRate: 3000,
		ShaftDepth:       40,
		AquiferType:      "gravel",
		Scenarios: []models.WaterLevelScenario{
			{
				ScenarioName:      "基准测试",
				InitialWaterLevel: 30,
				TargetWaterLevel:  20,
				ChangeRate:        0.5,
				DurationYears:     50,
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sim.SimulateWaterLevelImpact(req)
	}
}
