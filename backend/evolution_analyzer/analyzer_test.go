package evolutionanalyzer

import (
	"testing"
)

func TestNewEvolutionAnalyzer(t *testing.T) {
	analyzer := New()
	if analyzer == nil {
		t.Fatal("Expected non-nil EvolutionAnalyzer")
	}
}

func TestAnalyze_Normal(t *testing.T) {
	analyzer := New()
	result := analyzer.Analyze()

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if len(result.Evolutions) == 0 {
		t.Error("Expected non-empty evolutions list")
	}

	if len(result.KeyInnovations) == 0 {
		t.Error("Expected non-empty innovations list")
	}

	if result.Summary == "" {
		t.Error("Expected non-empty summary")
	}
}

func TestAnalyze_EraCount(t *testing.T) {
	analyzer := New()
	result := analyzer.Analyze()

	expectedEras := 7
	if len(result.Evolutions) != expectedEras {
		t.Errorf("Expected %d eras, got %d", expectedEras, len(result.Evolutions))
	}
}

func TestAnalyze_ProgressValidation(t *testing.T) {
	analyzer := New()
	result := analyzer.Analyze()

	prevDepth := 0.0
	prevFlowRate := 0.0
	prevLossRate := 100.0

	for _, era := range result.Evolutions {
		if era.AverageDepth < prevDepth {
			t.Errorf("Depth should increase over time: %s has depth %f, previous was %f",
				era.Era, era.AverageDepth, prevDepth)
		}
		prevDepth = era.AverageDepth

		if era.MaxFlowRate < prevFlowRate {
			t.Errorf("Flow rate should increase over time: %s has flow %f, previous was %f",
				era.Era, era.MaxFlowRate, prevFlowRate)
		}
		prevFlowRate = era.MaxFlowRate

		if era.WaterLossRate > prevLossRate {
			t.Errorf("Water loss rate should decrease over time: %s has loss %f, previous was %f",
				era.Era, era.WaterLossRate, prevLossRate)
		}
		prevLossRate = era.WaterLossRate
	}
}

func TestAnalyze_InnovationCount(t *testing.T) {
	analyzer := New()
	result := analyzer.Analyze()

	expectedInnovations := 8
	if len(result.KeyInnovations) != expectedInnovations {
		t.Errorf("Expected %d key innovations, got %d", expectedInnovations, len(result.KeyInnovations))
	}
}

func TestAnalyze_DataSourceAndReferences(t *testing.T) {
	analyzer := New()
	result := analyzer.Analyze()

	for _, era := range result.Evolutions {
		if era.DataSource == "" {
			t.Errorf("Era %s should have a data source", era.Era)
		}

		if len(era.References) < 2 {
			t.Errorf("Era %s should have at least 2 references, got %d",
				era.Era, len(era.References))
		}

		if era.ConfidenceLevel == "" {
			t.Errorf("Era %s should have a confidence level", era.Era)
		}
	}
}

func TestAnalyze_ResearchMethod(t *testing.T) {
	analyzer := New()
	result := analyzer.Analyze()

	if result.ResearchMethod == "" {
		t.Error("Expected non-empty research method")
	}

	if len(result.DataSources) == 0 {
		t.Error("Expected non-empty data sources list")
	}
}

func TestAnalyze_InnovationImpactRange(t *testing.T) {
	analyzer := New()
	result := analyzer.Analyze()

	for _, innovation := range result.KeyInnovations {
		if innovation.Impact < 0 || innovation.Impact > 10 {
			t.Errorf("Innovation %s impact should be in range [0, 10], got %f",
				innovation.Name, innovation.Impact)
		}
	}
}

func TestEraTechnology_Structure(t *testing.T) {
	analyzer := New()
	result := analyzer.Analyze()

	era := result.Evolutions[0]

	if era.Era == "" {
		t.Error("Era name should not be empty")
	}

	if era.TimePeriod == "" {
		t.Error("Time period should not be empty")
	}

	if len(era.KeyFeatures) == 0 {
		t.Error("Key features should not be empty")
	}

	if len(era.Materials) == 0 {
		t.Error("Materials should not be empty")
	}

	if len(era.ConstructionTools) == 0 {
		t.Error("Construction tools should not be empty")
	}
}
