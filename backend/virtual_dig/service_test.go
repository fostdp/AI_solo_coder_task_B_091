package virtualdig

import (
	"karez-system/models"
	"math"
	"sync"
	"testing"
	"time"
)

func TestNewVirtualDigService(t *testing.T) {
	service := New()
	if service == nil {
		t.Fatal("New() should return non-nil VirtualDigService")
	}
	if service.projects == nil {
		t.Error("projects map should be initialized")
	}
	if len(service.projects) != 0 {
		t.Error("projects should be empty initially")
	}
}

func TestGenerateID(t *testing.T) {
	service := New()

	id1 := service.generateID()
	id2 := service.generateID()

	if id1 == "" || id2 == "" {
		t.Error("Generated ID should not be empty")
	}
	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}
	if len(id1) != 16 {
		t.Errorf("ID should be 16 hex chars, got %d chars", len(id1))
	}
}

func TestGetDefaultTerrain(t *testing.T) {
	service := New()
	terrain := service.GetDefaultTerrain()

	if terrain.WidthKm <= 0 {
		t.Errorf("WidthKm should be positive, got %f", terrain.WidthKm)
	}
	if terrain.LengthKm <= 0 {
		t.Errorf("LengthKm should be positive, got %f", terrain.LengthKm)
	}
	if terrain.HeadElevation <= terrain.TailElevation {
		t.Error("HeadElevation should be > TailElevation")
	}
	if terrain.WaterTableDepth <= 0 {
		t.Errorf("WaterTableDepth should be positive, got %f", terrain.WaterTableDepth)
	}
	if terrain.SoilType == "" {
		t.Error("SoilType should not be empty")
	}
	if len(terrain.Obstacles) == 0 {
		t.Error("Should have some default obstacles")
	}

	for i, obs := range terrain.Obstacles {
		if obs.ID == "" {
			t.Errorf("Obstacle %d: ID should not be empty", i)
		}
		if obs.Type == "" {
			t.Errorf("Obstacle %d: Type should not be empty", i)
		}
		if obs.Radius <= 0 {
			t.Errorf("Obstacle %d: Radius should be positive, got %f", i, obs.Radius)
		}
	}
}

func TestSaveProject_Normal(t *testing.T) {
	service := New()

	terrain := service.GetDefaultTerrain()
	req := models.VirtualDigSaveRequest{
		ProjectName: "测试项目",
		Creator:     "测试用户",
		TerrainMap:  terrain,
		Channels: []models.DigChannel{
			{
				IsMain: true,
				Width:  1.2,
				Height: 1.8,
				Depth:  25,
				Points: []models.GeoPoint{
					{X: 1, Y: 0, Elevation: 840},
					{X: 1, Y: 2, Elevation: 830},
					{X: 1, Y: 4, Elevation: 820},
					{X: 1, Y: 6, Elevation: 810},
				},
			},
		},
		Shafts: []models.DigShaft{
			{
				ChannelID: "ch1",
				Position:  models.GeoPoint{X: 1, Y: 2, Elevation: 830},
				Depth:     35,
				Diameter:  1.2,
			},
			{
				ChannelID: "ch1",
				Position:  models.GeoPoint{X: 1, Y: 4, Elevation: 820},
				Depth:     35,
				Diameter:  1.2,
			},
		},
	}

	project, err := service.SaveProject(req)

	if err != nil {
		t.Fatalf("SaveProject failed: %v", err)
	}
	if project == nil {
		t.Fatal("SaveProject should return non-nil project")
	}
	if project.ID == "" {
		t.Error("Project ID should not be empty")
	}
	if project.ProjectName != "测试项目" {
		t.Errorf("Project name mismatch: expected '测试项目', got '%s'", project.ProjectName)
	}
	if project.Creator != "测试用户" {
		t.Errorf("Creator mismatch")
	}
	if project.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if len(project.Channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(project.Channels))
	}
	if len(project.Shafts) != 2 {
		t.Errorf("Expected 2 shafts, got %d", len(project.Shafts))
	}

	ch := project.Channels[0]
	if ch.Length <= 0 {
		t.Error("Channel length should be calculated and > 0")
	}
	if ch.Slope <= 0 {
		t.Error("Channel slope should be positive")
	}

	if project.Statistics.TotalChannelLength != ch.Length {
		t.Error("Statistics.TotalChannelLength mismatch")
	}
	if project.Statistics.TotalShafts != 2 {
		t.Errorf("Expected 2 shafts in stats, got %d", project.Statistics.TotalShafts)
	}
	if project.Statistics.TotalExcavationVolume <= 0 {
		t.Error("TotalExcavationVolume should be positive")
	}
	if project.Statistics.EstimatedCost <= 0 {
		t.Error("EstimatedCost should be positive")
	}

	if project.SimulatedFlow <= 0 {
		t.Error("SimulatedFlow should be positive")
	}

	if project.Feasibility.OverallScore <= 0 || project.Feasibility.OverallScore > 100 {
		t.Errorf("OverallScore should be between 0 and 100, got %f", project.Feasibility.OverallScore)
	}

	if !project.Feasibility.IsFeasible {
		t.Log("Warning: test project marked as not feasible, may be acceptable")
	}

	_, exists := service.projects[project.ID]
	if !exists {
		t.Error("Project should be saved to internal map")
	}
}

func TestSaveProject_AutoFillIDs(t *testing.T) {
	service := New()

	terrain := service.GetDefaultTerrain()
	req := models.VirtualDigSaveRequest{
		ProjectName: "自动填充测试",
		Creator:     "用户",
		TerrainMap:  terrain,
		Channels: []models.DigChannel{
			{
				Points: []models.GeoPoint{
					{X: 1, Y: 0, Elevation: 840},
					{X: 1, Y: 5, Elevation: 815},
				},
			},
		},
		Shafts: []models.DigShaft{
			{
				Position: models.GeoPoint{X: 1, Y: 2.5, Elevation: 827},
			},
		},
	}

	project, err := service.SaveProject(req)
	if err != nil {
		t.Fatalf("SaveProject failed: %v", err)
	}

	if project.Channels[0].ID == "" {
		t.Error("Channel ID should be auto-filled")
	}
	if project.Channels[0].Name == "" {
		t.Error("Channel name should be auto-filled")
	}
	if project.Shafts[0].ID == "" {
		t.Error("Shaft ID should be auto-filled")
	}
	if project.Shafts[0].Name == "" {
		t.Error("Shaft name should be auto-filled")
	}
}

func TestSaveProject_Minimal(t *testing.T) {
	service := New()

	terrain := service.GetDefaultTerrain()
	req := models.VirtualDigSaveRequest{
		ProjectName: "最小项目",
		TerrainMap:  terrain,
	}

	project, err := service.SaveProject(req)
	if err != nil {
		t.Fatalf("SaveProject failed: %v", err)
	}

	if len(project.Channels) != 0 {
		t.Error("Should have 0 channels")
	}
	if len(project.Shafts) != 0 {
		t.Error("Should have 0 shafts")
	}
	if project.SimulatedFlow != 0 {
		t.Error("SimulatedFlow should be 0 with no channels")
	}
	if project.Statistics.TotalChannelLength != 0 {
		t.Error("TotalChannelLength should be 0")
	}
}

func TestSimulateDesign(t *testing.T) {
	service := New()

	terrain := service.GetDefaultTerrain()
	req := models.VirtualDigSimulateRequest{
		TerrainMap: terrain,
		Channels: []models.DigChannel{
			{
				IsMain: true,
				Width:  1.0,
				Height: 1.5,
				Points: []models.GeoPoint{
					{X: 2, Y: 0, Elevation: 845},
					{X: 2, Y: 3, Elevation: 830},
					{X: 2, Y: 6, Elevation: 815},
				},
			},
		},
		Shafts: []models.DigShaft{
			{
				Position: models.GeoPoint{X: 2, Y: 3, Elevation: 830},
				Depth:    40,
				Diameter: 1.0,
			},
		},
	}

	project, err := service.SimulateDesign(req)
	if err != nil {
		t.Fatalf("SimulateDesign failed: %v", err)
	}
	if project == nil {
		t.Fatal("SimulateDesign should return non-nil project")
	}
	if project.ID == "" {
		t.Error("Project ID should be set")
	}
	if project.SimulatedFlow <= 0 {
		t.Error("Should have simulated flow")
	}
	if project.Feasibility.OverallScore <= 0 {
		t.Error("Should have feasibility score")
	}

	_, exists := service.projects[project.ID]
	if !exists {
		t.Error("Simulated project should be saved")
	}
}

func TestGetProject(t *testing.T) {
	service := New()

	terrain := service.GetDefaultTerrain()
	req := models.VirtualDigSaveRequest{
		ProjectName: "查询测试",
		TerrainMap:  terrain,
	}
	saved, _ := service.SaveProject(req)

	t.Run("existing_project", func(t *testing.T) {
		got, exists := service.GetProject(saved.ID)
		if !exists {
			t.Fatal("Should find existing project")
		}
		if got.ID != saved.ID {
			t.Errorf("ID mismatch: %s vs %s", got.ID, saved.ID)
		}
		if got.ProjectName != saved.ProjectName {
			t.Error("ProjectName mismatch")
		}
	})

	t.Run("non_existing_project", func(t *testing.T) {
		_, exists := service.GetProject("non_existent_id")
		if exists {
			t.Error("Should not find non-existing project")
		}
	})
}

func TestListProjects(t *testing.T) {
	service := New()

	terrain := service.GetDefaultTerrain()

	t.Run("empty_list", func(t *testing.T) {
		projects := service.ListProjects()
		if len(projects) != 0 {
			t.Errorf("Expected 0 projects, got %d", len(projects))
		}
	})

	t.Run("with_projects", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			req := models.VirtualDigSaveRequest{
				ProjectName: "项目" + intToStr(i),
				TerrainMap:  terrain,
			}
			service.SaveProject(req)
		}

		projects := service.ListProjects()
		if len(projects) != 3 {
			t.Errorf("Expected 3 projects, got %d", len(projects))
		}
	})
}

func TestDeleteProject(t *testing.T) {
	service := New()

	terrain := service.GetDefaultTerrain()
	req := models.VirtualDigSaveRequest{
		ProjectName: "删除测试",
		TerrainMap:  terrain,
	}
	saved, _ := service.SaveProject(req)

	t.Run("existing_project", func(t *testing.T) {
		ok := service.DeleteProject(saved.ID)
		if !ok {
			t.Error("Delete should return true for existing project")
		}
		_, exists := service.projects[saved.ID]
		if exists {
			t.Error("Project should be removed from map")
		}
	})

	t.Run("non_existing_project", func(t *testing.T) {
		ok := service.DeleteProject("non_existent")
		if ok {
			t.Error("Delete should return false for non-existing project")
		}
	})
}

func TestFeasibilityEvaluation_Hydraulics(t *testing.T) {
	service := New()

	t.Run("no_main_channel", func(t *testing.T) {
		terrain := service.GetDefaultTerrain()
		req := models.VirtualDigSaveRequest{
			ProjectName: "无主渠测试",
			TerrainMap:  terrain,
			Channels: []models.DigChannel{
				{
					IsMain: false,
					Points: []models.GeoPoint{
						{X: 1, Y: 0, Elevation: 840},
						{X: 1, Y: 3, Elevation: 825},
					},
				},
			},
		}
		project, _ := service.SaveProject(req)

		hasNoMainIssue := false
		for _, issue := range project.Feasibility.Issues {
			if issue.Severity == "严重" && contains(issue.Message, "缺少主暗渠") {
				hasNoMainIssue = true
				break
			}
		}
		if !hasNoMainIssue {
			t.Error("Should flag missing main channel as serious issue")
		}
		if project.Feasibility.HydraulicScore >= 70 {
			t.Errorf("Hydraulic score should be low without main channel, got %f", project.Feasibility.HydraulicScore)
		}
	})

	t.Run("zero_slope", func(t *testing.T) {
		terrain := service.GetDefaultTerrain()
		req := models.VirtualDigSaveRequest{
			ProjectName: "零坡度测试",
			TerrainMap:  terrain,
			Channels: []models.DigChannel{
				{
					IsMain: true,
					Points: []models.GeoPoint{
						{X: 1, Y: 0, Elevation: 840},
						{X: 1, Y: 3, Elevation: 840},
						{X: 1, Y: 6, Elevation: 840},
					},
				},
			},
		}
		project, _ := service.SaveProject(req)

		hasSlopeIssue := false
		for _, issue := range project.Feasibility.Issues {
			if contains(issue.Message, "坡度") {
				hasSlopeIssue = true
				break
			}
		}
		if !hasSlopeIssue {
			t.Error("Should flag slope issues")
		}
	})

	t.Run("negative_slope", func(t *testing.T) {
		terrain := service.GetDefaultTerrain()
		req := models.VirtualDigSaveRequest{
			ProjectName: "负坡度测试",
			TerrainMap:  terrain,
			Channels: []models.DigChannel{
				{
					IsMain: true,
					Points: []models.GeoPoint{
						{X: 1, Y: 0, Elevation: 810},
						{X: 1, Y: 3, Elevation: 825},
						{X: 1, Y: 6, Elevation: 840},
					},
				},
			},
		}
		project, _ := service.SaveProject(req)

		if project.Feasibility.HydraulicScore >= 60 {
			t.Errorf("Hydraulic score should be low with negative slope, got %f", project.Feasibility.HydraulicScore)
		}
	})
}

func TestFeasibilityEvaluation_Geology(t *testing.T) {
	service := New()

	t.Run("good_soil_gravel", func(t *testing.T) {
		terrain := service.GetDefaultTerrain()
		terrain.SoilType = "gravel"
		terrain.Obstacles = nil
		req := models.VirtualDigSaveRequest{
			ProjectName: "好土质测试",
			TerrainMap:  terrain,
			Channels: []models.DigChannel{
				{
					IsMain: true,
					Points: []models.GeoPoint{
						{X: 1, Y: 0, Elevation: 840},
						{X: 1, Y: 3, Elevation: 825},
						{X: 1, Y: 6, Elevation: 810},
					},
				},
			},
		}
		project, _ := service.SaveProject(req)

		if project.Feasibility.GeologicalScore < 70 {
			t.Errorf("Geological score should be high for gravel, got %f", project.Feasibility.GeologicalScore)
		}
	})

	t.Run("poor_soil_clay", func(t *testing.T) {
		terrain := service.GetDefaultTerrain()
		terrain.SoilType = "clay"
		terrain.Obstacles = nil
		req := models.VirtualDigSaveRequest{
			ProjectName: "黏土测试",
			TerrainMap:  terrain,
			Channels: []models.DigChannel{
				{
					IsMain: true,
					Points: []models.GeoPoint{
						{X: 1, Y: 0, Elevation: 840},
						{X: 1, Y: 3, Elevation: 825},
						{X: 1, Y: 6, Elevation: 810},
					},
				},
			},
		}
		project, _ := service.SaveProject(req)

		hasClayIssue := false
		for _, issue := range project.Feasibility.Issues {
			if contains(issue.Message, "黏土") {
				hasClayIssue = true
				break
			}
		}
		if !hasClayIssue {
			t.Error("Should flag clay soil issues")
		}
	})

	t.Run("obstacle_hit", func(t *testing.T) {
		terrain := service.GetDefaultTerrain()
		terrain.Obstacles = []models.MapObstacle{
			{
				ID:     "rock1",
				Type:   "rock",
				X:      1,
				Y:      3,
				Radius: 0.5,
				Label:  "花岗岩体",
			},
		}
		req := models.VirtualDigSaveRequest{
			ProjectName: "障碍物测试",
			TerrainMap:  terrain,
			Channels: []models.DigChannel{
				{
					IsMain: true,
					Points: []models.GeoPoint{
						{X: 1, Y: 0, Elevation: 840},
						{X: 1, Y: 3, Elevation: 825},
						{X: 1, Y: 6, Elevation: 810},
					},
				},
			},
		}
		project, _ := service.SaveProject(req)

		hasObstacleIssue := false
		for _, issue := range project.Feasibility.Issues {
			if contains(issue.Message, "花岗岩体") {
				hasObstacleIssue = true
				break
			}
		}
		if !hasObstacleIssue {
			t.Error("Should flag obstacle collision")
		}
	})
}

func TestCalculateChannelMetrics(t *testing.T) {
	service := New()
	terrain := service.GetDefaultTerrain()

	t.Run("valid_channel", func(t *testing.T) {
		ch := &models.DigChannel{
			Points: []models.GeoPoint{
				{X: 0, Y: 0, Elevation: 850},
				{X: 0, Y: 1, Elevation: 845},
				{X: 0, Y: 2, Elevation: 840},
				{X: 0, Y: 3, Elevation: 835},
			},
		}
		service.calculateChannelMetrics(ch, terrain)

		expectedLength := 3000.0
		if math.Abs(ch.Length-expectedLength) > 1 {
			t.Errorf("Expected length ~%f, got %f", expectedLength, ch.Length)
		}

		expectedSlope := 15.0 / 3000.0
		if math.Abs(ch.Slope-expectedSlope) > 0.0001 {
			t.Errorf("Expected slope ~%f, got %f", expectedSlope, ch.Slope)
		}
	})

	t.Run("single_point", func(t *testing.T) {
		ch := &models.DigChannel{
			Points: []models.GeoPoint{
				{X: 0, Y: 0, Elevation: 850},
			},
		}
		service.calculateChannelMetrics(ch, terrain)

		if ch.Length != 0 {
			t.Errorf("Expected length 0 for single point, got %f", ch.Length)
		}
		if ch.Slope != 0 {
			t.Errorf("Expected slope 0 for single point, got %f", ch.Slope)
		}
	})

	t.Run("auto_fill_dimensions", func(t *testing.T) {
		ch := &models.DigChannel{
			Points: []models.GeoPoint{
				{X: 0, Y: 0, Elevation: 850},
				{X: 0, Y: 3, Elevation: 835},
			},
		}
		service.calculateChannelMetrics(ch, terrain)

		if ch.Width == 0 {
			t.Error("Width should be auto-filled")
		}
		if ch.Height == 0 {
			t.Error("Height should be auto-filled")
		}
		if ch.Depth == 0 {
			t.Error("Depth should be auto-filled")
		}
	})
}

func TestCalculateShaftMetrics(t *testing.T) {
	service := New()
	terrain := service.GetDefaultTerrain()

	t.Run("reaches_water", func(t *testing.T) {
		sh := &models.DigShaft{
			Position: models.GeoPoint{X: 2, Y: 4, Elevation: 820},
		}
		sh.Depth = 60
		service.calculateShaftMetrics(sh, terrain)

		if !sh.ReachesWater {
			t.Error("Deep shaft should reach water table")
		}
	})

	t.Run("does_not_reach_water", func(t *testing.T) {
		sh := &models.DigShaft{
			Position: models.GeoPoint{X: 2, Y: 4, Elevation: 820},
		}
		sh.Depth = 5
		service.calculateShaftMetrics(sh, terrain)

		if sh.ReachesWater {
			t.Error("Shallow shaft should not reach water table")
		}
	})

	t.Run("auto_fill", func(t *testing.T) {
		sh := &models.DigShaft{
			Position: models.GeoPoint{X: 2, Y: 4, Elevation: 820},
		}
		service.calculateShaftMetrics(sh, terrain)

		if sh.Depth == 0 {
			t.Error("Depth should be auto-filled")
		}
		if sh.Diameter == 0 {
			t.Error("Diameter should be auto-filled")
		}
		if sh.DistanceFromHead == 0 {
			t.Error("DistanceFromHead should be auto-filled")
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	service := New()
	terrain := service.GetDefaultTerrain()

	var wg sync.WaitGroup
	numOperations := 50

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			req := models.VirtualDigSaveRequest{
				ProjectName: "并发测试" + intToStr(idx),
				TerrainMap:  terrain,
				Channels: []models.DigChannel{
					{
						IsMain: true,
						Points: []models.GeoPoint{
							{X: float64(idx) * 0.1, Y: 0, Elevation: 840},
							{X: float64(idx) * 0.1, Y: 3, Elevation: 825},
						},
					},
				},
			}

			project, err := service.SaveProject(req)
			if err != nil {
				t.Errorf("Concurrent save failed: %v", err)
				return
			}

			_, exists := service.GetProject(project.ID)
			if !exists {
				t.Error("Concurrent get failed")
			}

			_ = service.ListProjects()
		}(i)
	}

	wg.Wait()

	projects := service.ListProjects()
	if len(projects) != numOperations {
		t.Errorf("Expected %d projects, got %d", numOperations, len(projects))
	}
}

func TestEstimateFlow_DifferentSoils(t *testing.T) {
	service := New()

	soils := []string{"gravel", "sand", "silt", "clay", "limestone"}
	flows := make(map[string]float64)

	channels := []models.DigChannel{
		{
			IsMain: true,
			Width:  1.2,
			Height: 1.8,
			Points: []models.GeoPoint{
				{X: 1, Y: 0, Elevation: 840},
				{X: 1, Y: 3, Elevation: 825},
				{X: 1, Y: 6, Elevation: 810},
			},
		},
	}

	shafts := []models.DigShaft{
		{
			Position:     models.GeoPoint{X: 1, Y: 3, Elevation: 825},
			Depth:        40,
			Diameter:     1.2,
			ReachesWater: true,
		},
	}

	for _, soil := range soils {
		terrain := service.GetDefaultTerrain()
		terrain.SoilType = soil

		for i := range channels {
			service.calculateChannelMetrics(&channels[i], terrain)
		}

		flow := service.estimateFlow(channels, shafts, terrain)
		flows[soil] = flow

		if flow < 0 {
			t.Errorf("Flow should not be negative for %s, got %f", soil, flow)
		}
	}

	if flows["gravel"] <= flows["sand"] {
		t.Error("Gravel should have higher flow than sand")
	}
	if flows["sand"] <= flows["silt"] {
		t.Error("Sand should have higher flow than silt")
	}
	if flows["silt"] <= flows["clay"] {
		t.Error("Silt should have higher flow than clay")
	}
}

func TestGetSoilFlowFactor(t *testing.T) {
	service := New()

	testCases := []struct {
		soil string
		min  float64
		max  float64
	}{
		{"gravel", 1.1, 1.3},
		{"sand", 0.9, 1.1},
		{"silt", 0.6, 0.8},
		{"clay", 0.3, 0.5},
		{"limestone", 0.8, 1.0},
		{"unknown", 0.9, 1.1},
	}

	for _, tc := range testCases {
		factor := service.getSoilFlowFactor(tc.soil)
		if factor < tc.min || factor > tc.max {
			t.Errorf("%s: factor %f out of expected range [%f, %f]", tc.soil, factor, tc.min, tc.max)
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func BenchmarkSaveProject(b *testing.B) {
	service := New()
	terrain := service.GetDefaultTerrain()
	req := models.VirtualDigSaveRequest{
		ProjectName: "基准测试项目",
		Creator:     "benchmark",
		TerrainMap:  terrain,
		Channels: []models.DigChannel{
			{
				IsMain: true,
				Width:  1.2,
				Height: 1.8,
				Points: []models.GeoPoint{
					{X: 1, Y: 0, Elevation: 840},
					{X: 1, Y: 2, Elevation: 830},
					{X: 1, Y: 4, Elevation: 820},
					{X: 1, Y: 6, Elevation: 810},
				},
			},
		},
		Shafts: []models.DigShaft{
			{Position: models.GeoPoint{X: 1, Y: 2, Elevation: 830}, Depth: 35, Diameter: 1.2},
			{Position: models.GeoPoint{X: 1, Y: 4, Elevation: 820}, Depth: 35, Diameter: 1.2},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.SaveProject(req)
	}
}

func BenchmarkSimulateDesign(b *testing.B) {
	service := New()
	terrain := service.GetDefaultTerrain()
	req := models.VirtualDigSimulateRequest{
		TerrainMap: terrain,
		Channels: []models.DigChannel{
			{
				IsMain: true,
				Width:  1.2,
				Height: 1.8,
				Points: []models.GeoPoint{
					{X: 1, Y: 0, Elevation: 840},
					{X: 1, Y: 3, Elevation: 825},
					{X: 1, Y: 6, Elevation: 810},
				},
			},
		},
		Shafts: []models.DigShaft{
			{Position: models.GeoPoint{X: 1, Y: 3, Elevation: 825}, Depth: 40, Diameter: 1.2},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.SimulateDesign(req)
	}
}

func TestSaveProject_Issues(t *testing.T) {
	service := New()

	t.Run("insufficient_shafts", func(t *testing.T) {
		terrain := service.GetDefaultTerrain()
		terrain.Obstacles = nil
		req := models.VirtualDigSaveRequest{
			ProjectName: "竖井不足测试",
			TerrainMap:  terrain,
			Channels: []models.DigChannel{
				{
					IsMain: true,
					Points: []models.GeoPoint{
						{X: 1, Y: 0, Elevation: 840},
						{X: 1, Y: 3, Elevation: 825},
						{X: 1, Y: 6, Elevation: 810},
					},
				},
			},
			Shafts: []models.DigShaft{
				{Position: models.GeoPoint{X: 1, Y: 3, Elevation: 825}, Depth: 5, Diameter: 1.2},
			},
		}
		project, _ := service.SaveProject(req)

		hasInsufficientShafts := false
		hasDryShafts := false
		for _, issue := range project.Feasibility.Issues {
			if contains(issue.Message, "竖井数量不足") {
				hasInsufficientShafts = true
			}
			if contains(issue.Message, "未触及地下水位") {
				hasDryShafts = true
			}
		}
		if !hasInsufficientShafts {
			t.Error("Should flag insufficient shafts")
		}
		if !hasDryShafts {
			t.Error("Should flag shafts not reaching water")
		}
		if project.Feasibility.IsFeasible {
			t.Error("Project with issues should not be feasible")
		}
	})

	t.Run("excessive_cost", func(t *testing.T) {
		terrain := service.GetDefaultTerrain()
		terrain.Obstacles = nil
		manyPoints := make([]models.GeoPoint, 0, 200)
		for i := 0; i < 200; i++ {
			manyPoints = append(manyPoints, models.GeoPoint{
				X:         1,
				Y:         float64(i) * 0.1,
				Elevation: 840 - float64(i)*0.25,
			})
		}

		shafts := make([]models.DigShaft, 0, 50)
		for i := 0; i < 50; i++ {
			shafts = append(shafts, models.DigShaft{
				Position: models.GeoPoint{X: 1, Y: float64(i) * 0.4, Elevation: 830},
				Depth:    50,
				Diameter: 2.0,
			})
		}

		req := models.VirtualDigSaveRequest{
			ProjectName: "高成本测试",
			TerrainMap:  terrain,
			Channels: []models.DigChannel{
				{
					IsMain: true,
					Width:  5,
					Height: 5,
					Points: manyPoints,
				},
			},
			Shafts: shafts,
		}
		project, _ := service.SaveProject(req)

		if project.Statistics.TotalExcavationVolume <= 100000 {
			t.Errorf("Volume should be > 100000, got %f", project.Statistics.TotalExcavationVolume)
		}

		hasCostIssue := false
		for _, issue := range project.Feasibility.Issues {
			if contains(issue.Message, "工程量较大") {
				hasCostIssue = true
				break
			}
		}
		if !hasCostIssue {
			t.Error("Should flag large project volume")
		}
	})
}

func TestCalculateStatistics(t *testing.T) {
	service := New()

	channels := []models.DigChannel{
		{
			Length: 1000,
			Width:  1.0,
			Height: 1.5,
			Depth:  20,
		},
		{
			Length: 500,
			Width:  0.8,
			Height: 1.2,
			Depth:  15,
		},
	}

	shafts := []models.DigShaft{
		{
			Depth:    30,
			Diameter: 1.0,
		},
		{
			Depth:    25,
			Diameter: 1.2,
		},
	}

	stats := service.calculateStatistics(channels, shafts)

	expectedLength := 1500.0
	if math.Abs(stats.TotalChannelLength-expectedLength) > 0.01 {
		t.Errorf("Expected length %f, got %f", expectedLength, stats.TotalChannelLength)
	}

	if stats.TotalShafts != 2 {
		t.Errorf("Expected 2 shafts, got %d", stats.TotalShafts)
	}

	channelVolume := 1000*1.0*1.5 + 500*0.8*1.2
	shaft1Volume := math.Pi * 0.5 * 0.5 * 30
	shaft2Volume := math.Pi * 0.6 * 0.6 * 25
	expectedVolume := channelVolume + shaft1Volume + shaft2Volume

	if math.Abs(stats.TotalExcavationVolume-expectedVolume) > 0.01 {
		t.Errorf("Expected volume ~%f, got %f", expectedVolume, stats.TotalExcavationVolume)
	}

	if stats.EstimatedManDays <= 0 {
		t.Error("EstimatedManDays should be positive")
	}
	if stats.EstimatedCost <= 0 {
		t.Error("EstimatedCost should be positive")
	}
	if stats.AverageDepth <= 0 {
		t.Error("AverageDepth should be positive")
	}
}

func TestTimeConsistency(t *testing.T) {
	service := New()
	terrain := service.GetDefaultTerrain()

	before := time.Now()
	time.Sleep(1 * time.Millisecond)

	project, _ := service.SaveProject(models.VirtualDigSaveRequest{
		ProjectName: "时间测试",
		TerrainMap:  terrain,
	})

	if project.CreatedAt.Before(before) {
		t.Error("CreatedAt should be after 'before' time")
	}
	if project.CreatedAt.After(time.Now()) {
		t.Error("CreatedAt should be before now")
	}
}
