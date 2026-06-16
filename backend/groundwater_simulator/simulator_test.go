package groundwatersimulator

import (
	"context"
	"testing"
	"time"

	"karez-system/models"
)

func TestNewGroundwaterSimulator(t *testing.T) {
	sim := New()
	if sim == nil {
		t.Fatal("Expected non-nil GroundwaterSimulator")
	}
}

func TestGetAquiferInfo_ValidTypes(t *testing.T) {
	sim := New()

	validTypes := []string{"gravel", "sand", "silt", "clay", "limestone"}

	for _, aquiferType := range validTypes {
		t.Run(aquiferType, func(t *testing.T) {
			info := sim.GetAquiferInfo(aquiferType)
			if info.Type != aquiferType {
				t.Errorf("Expected type %s, got %s", aquiferType, info.Type)
			}
			if info.Permeability <= 0 {
				t.Error("Permeability should be positive")
			}
			if info.DataSource == "" {
				t.Error("DataSource should not be empty")
			}
			if info.TypicalLocations == "" {
				t.Error("TypicalLocations should not be empty")
			}
		})
	}
}

func TestGetAquiferInfo_InvalidType(t *testing.T) {
	sim := New()
	info := sim.GetAquiferInfo("invalid_type")

	if info.Type != "invalid_type" {
		t.Error("Should return info with the requested type name")
	}

	defaultInfo := sim.GetAquiferInfo("gravel")
	if info.Permeability != defaultInfo.Permeability {
		t.Error("Invalid type should default to gravel parameters")
	}
}

func TestGetDefaultScenarios(t *testing.T) {
	sim := New()
	scenarios := sim.GetDefaultScenarios()

	expectedCount := 5
	if len(scenarios) != expectedCount {
		t.Errorf("Expected %d scenarios, got %d", expectedCount, len(scenarios))
	}

	for _, scenario := range scenarios {
		if scenario.ScenarioName == "" {
			t.Error("Scenario name should not be empty")
		}
		if scenario.Description == "" {
			t.Error("Scenario description should not be empty")
		}
	}
}

func TestSimulate_StableScenario(t *testing.T) {
	sim := New()
	req := models.WaterLevelSimulationRequest{
		KarezID:          1,
		BaselineFlowRate: 3000,
		ShaftDepth:       40,
		AquiferType:      "gravel",
		Scenarios: []models.WaterLevelScenario{
			{
				ScenarioName:      "稳定状态",
				InitialWaterLevel: 30,
				TargetWaterLevel:  30,
				ChangeRate:        0,
				DurationYears:     50,
			},
		},
	}

	results := sim.Simulate(req)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]

	if result.ScenarioName != "稳定状态" {
		t.Errorf("Expected scenario name '稳定状态', got %s", result.ScenarioName)
	}

	if len(result.DataPoints) == 0 {
		t.Error("Expected non-empty data points")
	}

	if result.FinalFlowRate <= 0 {
		t.Error("Final flow rate should be positive")
	}

	if result.BaselineFlow != 3000 {
		t.Errorf("Expected baseline flow 3000, got %f", result.BaselineFlow)
	}
}

func TestSimulate_DeclineScenario(t *testing.T) {
	sim := New()
	req := models.WaterLevelSimulationRequest{
		BaselineFlowRate: 3000,
		ShaftDepth:       40,
		Scenarios: []models.WaterLevelScenario{
			{
				ScenarioName:      "中度下降",
				InitialWaterLevel: 30,
				TargetWaterLevel:  10,
				ChangeRate:        0.5,
				DurationYears:     40,
			},
		},
	}

	results := sim.Simulate(req)
	result := results[0]

	firstFlow := result.DataPoints[0].FlowRate
	lastFlow := result.DataPoints[len(result.DataPoints)-1].FlowRate

	if lastFlow >= firstFlow {
		t.Error("Flow rate should decrease in decline scenario")
	}

	if result.TotalDecline <= 0 {
		t.Error("Total decline should be positive for decline scenario")
	}
}

func TestSimulate_RecoveryScenario(t *testing.T) {
	sim := New()
	req := models.WaterLevelSimulationRequest{
		BaselineFlowRate: 3000,
		ShaftDepth:       40,
		Scenarios: []models.WaterLevelScenario{
			{
				ScenarioName:      "生态恢复",
				InitialWaterLevel: 15,
				TargetWaterLevel:  30,
				ChangeRate:        -0.3,
				DurationYears:     50,
			},
		},
	}

	results := sim.Simulate(req)
	result := results[0]

	firstFlow := result.DataPoints[0].FlowRate
	lastFlow := result.DataPoints[len(result.DataPoints)-1].FlowRate

	if lastFlow <= firstFlow {
		t.Error("Flow rate should increase in recovery scenario")
	}
}

func TestSimulate_DefaultValues(t *testing.T) {
	sim := New()
	req := models.WaterLevelSimulationRequest{}

	results := sim.Simulate(req)

	if len(results) == 0 {
		t.Fatal("Expected results with default scenarios")
	}

	if results[0].BaselineFlow == 0 {
		t.Error("Baseline flow should have default value")
	}
}

func TestSimulate_EdgeCases(t *testing.T) {
	sim := New()

	testCases := []struct {
		name string
		req  models.WaterLevelSimulationRequest
	}{
		{
			name: "zero water level",
			req: models.WaterLevelSimulationRequest{
				BaselineFlowRate: 3000,
				ShaftDepth:       40,
				Scenarios: []models.WaterLevelScenario{
					{
						ScenarioName:      "Zero",
						InitialWaterLevel: 0,
						TargetWaterLevel:  0,
						ChangeRate:        0,
						DurationYears:     5,
					},
				},
			},
		},
		{
			name: "zero baseline flow",
			req: models.WaterLevelSimulationRequest{
				BaselineFlowRate: 0,
				ShaftDepth:       40,
				Scenarios: []models.WaterLevelScenario{
					{
						ScenarioName:      "ZeroFlow",
						InitialWaterLevel: 30,
						TargetWaterLevel:  30,
						ChangeRate:        0,
						DurationYears:     5,
					},
				},
			},
		},
		{
			name: "different aquifer types",
			req: models.WaterLevelSimulationRequest{
				BaselineFlowRate: 3000,
				ShaftDepth:       40,
				AquiferType:      "sand",
				Scenarios: []models.WaterLevelScenario{
					{
						ScenarioName:      "SandTest",
						InitialWaterLevel: 30,
						TargetWaterLevel:  30,
						ChangeRate:        0,
						DurationYears:     5,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results := sim.Simulate(tc.req)
			if len(results) == 0 {
				t.Fatalf("Expected results for case: %s", tc.name)
			}

			for _, result := range results {
				if len(result.DataPoints) == 0 {
					t.Errorf("Expected data points for case: %s", tc.name)
				}

				if result.FinalFlowRate < 0 {
					t.Errorf("Flow rate should not be negative for case: %s", tc.name)
				}
			}
		})
	}
}

func TestSimulate_FlowValidation(t *testing.T) {
	sim := New()
	req := models.WaterLevelSimulationRequest{
		BaselineFlowRate: 3000,
		ShaftDepth:       40,
		AquiferType:      "gravel",
		Scenarios: []models.WaterLevelScenario{
			{
				ScenarioName:      "Test",
				InitialWaterLevel: 30,
				TargetWaterLevel:  30,
				ChangeRate:        0,
				DurationYears:     10,
			},
		},
	}

	results := sim.Simulate(req)
	result := results[0]

	for _, dp := range result.DataPoints {
		if dp.FlowRate < 0 {
			t.Error("Flow rate should not be negative")
		}
		if dp.WaterLevel < 0 {
			t.Error("Water level should not be negative")
		}
	}
}

func TestSubmitAsync(t *testing.T) {
	sim := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sim.Start(ctx)

	req := models.WaterLevelSimulationRequest{
		KarezID:          1,
		BaselineFlowRate: 3000,
		ShaftDepth:       40,
	}

	taskID := sim.SubmitAsync(req)

	if taskID == "" {
		t.Fatal("Expected non-empty task ID")
	}

	task, exists := sim.GetTaskStatus(taskID)
	if !exists {
		t.Fatal("Expected task to exist")
	}

	if task.Status != "pending" && task.Status != "running" && task.Status != "completed" {
		t.Errorf("Unexpected task status: %s", task.Status)
	}
}

func TestGetTaskStatus_NonExistent(t *testing.T) {
	sim := New()

	task, exists := sim.GetTaskStatus("non-existent-task")
	if exists {
		t.Error("Should not find non-existent task")
	}
	if task != nil {
		t.Error("Task should be nil for non-existent task")
	}
}

func TestAsyncTaskCompletes(t *testing.T) {
	sim := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sim.Start(ctx)

	req := models.WaterLevelSimulationRequest{
		KarezID:          1,
		BaselineFlowRate: 3000,
		ShaftDepth:       40,
		Scenarios: []models.WaterLevelScenario{
			{
				ScenarioName:      "快速测试",
				InitialWaterLevel: 30,
				TargetWaterLevel:  30,
				ChangeRate:        0,
				DurationYears:     5,
			},
		},
	}

	taskID := sim.SubmitAsync(req)

	completed := false
	for i := 0; i < 20; i++ {
		time.Sleep(10 * time.Millisecond)
		task, exists := sim.GetTaskStatus(taskID)
		if !exists {
			t.Fatal("Task should exist")
		}
		if task.Status == "completed" {
			completed = true
			break
		}
	}

	if !completed {
		t.Error("Async task should complete within reasonable time")
	}

	task, _ := sim.GetTaskStatus(taskID)
	if len(task.Results) == 0 {
		t.Error("Completed task should have results")
	}
}

func TestAquiferInfo_ModelAssumptions(t *testing.T) {
	sim := New()

	req := models.WaterLevelSimulationRequest{
		BaselineFlowRate: 3000,
		ShaftDepth:       40,
		Scenarios: []models.WaterLevelScenario{
			{
				ScenarioName:      "Test",
				InitialWaterLevel: 30,
				TargetWaterLevel:  30,
				ChangeRate:        0,
				DurationYears:     5,
			},
		},
	}

	results := sim.Simulate(req)
	result := results[0]

	if len(result.ModelAssumptions) == 0 {
		t.Error("Expected non-empty model assumptions")
	}

	if result.AquiferInfo.DataSource == "" {
		t.Error("Expected aquifer info with data source")
	}
}

func TestStartWithContextCancellation(t *testing.T) {
	sim := New()
	ctx, cancel := context.WithCancel(context.Background())

	sim.Start(ctx)

	time.Sleep(10 * time.Millisecond)

	cancel()

	time.Sleep(20 * time.Millisecond)
}
