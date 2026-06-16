package vrkarezbuilder

import (
	"testing"

	"karez-system/models"
)

func TestNewVrKarezBuilder(t *testing.T) {
	builder := New()
	if builder == nil {
		t.Fatal("Expected non-nil VrKarezBuilder")
	}
}

func TestGenerateID(t *testing.T) {
	builder := New()
	id := builder.generateID()

	if id == "" {
		t.Error("Expected non-empty ID")
	}

	if len(id) != 16 {
		t.Errorf("Expected 16-character hex ID, got %d chars", len(id))
	}

	id2 := builder.generateID()
	if id == id2 {
		t.Error("Expected unique IDs")
	}
}

func TestGetDefaultTerrain(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	if terrain.WidthKm <= 0 {
		t.Error("Width should be positive")
	}

	if terrain.LengthKm <= 0 {
		t.Error("Length should be positive")
	}

	if terrain.HeadElevation <= terrain.TailElevation {
		t.Error("Head elevation should be higher than tail elevation")
	}

	if len(terrain.Obstacles) == 0 {
		t.Error("Expected some obstacles in default terrain")
	}
}

func TestSaveProject_Normal(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	req := models.VirtualDigSaveRequest{
		ProjectName: "Test Project",
		Creator:     "test_user",
		TerrainMap:  terrain,
		Channels: []models.DigChannel{
			{
				Name:   "主暗渠",
				IsMain: true,
				Width:  1.5,
				Height: 2.0,
				Points: []models.GeoPoint{
					{X: 1.0, Y: 0.5, Elevation: 845},
					{X: 1.0, Y: 4.0, Elevation: 820},
					{X: 1.0, Y: 7.5, Elevation: 800},
				},
			},
		},
		Shafts: []models.DigShaft{
			{Name: "1号竖井", Position: models.GeoPoint{X: 1.0, Y: 1.0, Elevation: 840}, Depth: 45},
			{Name: "2号竖井", Position: models.GeoPoint{X: 1.0, Y: 3.0, Elevation: 830}, Depth: 40},
			{Name: "3号竖井", Position: models.GeoPoint{X: 1.0, Y: 5.0, Elevation: 815}, Depth: 35},
			{Name: "4号竖井", Position: models.GeoPoint{X: 1.0, Y: 7.0, Elevation: 802}, Depth: 28},
		},
	}

	project, err := builder.SaveProject(req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if project == nil {
		t.Fatal("Expected non-nil project")
	}

	if project.ID == "" {
		t.Error("Expected project to have an ID")
	}

	if project.ProjectName != "Test Project" {
		t.Errorf("Expected project name 'Test Project', got '%s'", project.ProjectName)
	}

	if len(project.Channels) == 0 {
		t.Error("Expected channels in project")
	}

	if project.Statistics.TotalChannelLength <= 0 {
		t.Error("Expected positive total channel length")
	}
}

func TestSaveProject_AutoFillIDs(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	req := models.VirtualDigSaveRequest{
		ProjectName: "Test",
		TerrainMap:  terrain,
		Channels: []models.DigChannel{
			{
				Points: []models.GeoPoint{
					{X: 1.0, Y: 0.5, Elevation: 845},
					{X: 1.0, Y: 7.5, Elevation: 800},
				},
			},
		},
		Shafts: []models.DigShaft{
			{Position: models.GeoPoint{X: 1.0, Y: 4.0, Elevation: 820}, Depth: 30},
		},
	}

	project, err := builder.SaveProject(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if project.Channels[0].ID == "" {
		t.Error("Expected channel ID to be auto-filled")
	}

	if project.Channels[0].Name == "" {
		t.Error("Expected channel name to be auto-filled")
	}

	if project.Shafts[0].ID == "" {
		t.Error("Expected shaft ID to be auto-filled")
	}

	if project.Shafts[0].Name == "" {
		t.Error("Expected shaft name to be auto-filled")
	}
}

func TestSaveProject_Minimal(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	req := models.VirtualDigSaveRequest{
		ProjectName: "Minimal",
		TerrainMap:  terrain,
	}

	project, err := builder.SaveProject(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if project == nil {
		t.Fatal("Expected non-nil project")
	}

	if len(project.Channels) != 0 {
		t.Error("Expected no channels")
	}

	if len(project.Shafts) != 0 {
		t.Error("Expected no shafts")
	}

	if project.SimulatedFlow != 0 {
		t.Error("Expected zero flow with no channels")
	}
}

func TestSimulateDesign(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	req := models.VirtualDigSimulateRequest{
		TerrainMap: terrain,
		Channels: []models.DigChannel{
			{
				IsMain: true,
				Width:  1.2,
				Height: 1.8,
				Points: []models.GeoPoint{
					{X: 2.0, Y: 0.5, Elevation: 845},
					{X: 2.0, Y: 7.5, Elevation: 805},
				},
			},
		},
		Shafts: []models.DigShaft{
			{Position: models.GeoPoint{X: 2.0, Y: 1.0, Elevation: 842}, Depth: 50},
			{Position: models.GeoPoint{X: 2.0, Y: 4.0, Elevation: 825}, Depth: 40},
			{Position: models.GeoPoint{X: 2.0, Y: 7.0, Elevation: 808}, Depth: 30},
		},
	}

	project, err := builder.SimulateDesign(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if project == nil {
		t.Fatal("Expected non-nil project")
	}

	if project.Feasibility.OverallScore <= 0 {
		t.Error("Expected positive feasibility score")
	}
}

func TestGetProject(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	t.Run("existing project", func(t *testing.T) {
		req := models.VirtualDigSaveRequest{
			ProjectName: "Test",
			TerrainMap:  terrain,
		}
		project, _ := builder.SaveProject(req)

		found, exists := builder.GetProject(project.ID)
		if !exists {
			t.Error("Expected project to exist")
		}
		if found.ID != project.ID {
			t.Errorf("Expected project ID %s, got %s", project.ID, found.ID)
		}
	})

	t.Run("non existing project", func(t *testing.T) {
		_, exists := builder.GetProject("non-existent")
		if exists {
			t.Error("Expected project not to exist")
		}
	})
}

func TestListProjects(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	t.Run("empty list", func(t *testing.T) {
		projects := builder.ListProjects()
		if len(projects) != 0 {
			t.Errorf("Expected empty list, got %d projects", len(projects))
		}
	})

	t.Run("with projects", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			req := models.VirtualDigSaveRequest{
				ProjectName: "Test",
				TerrainMap:  terrain,
			}
			builder.SaveProject(req)
		}

		projects := builder.ListProjects()
		if len(projects) != 3 {
			t.Errorf("Expected 3 projects, got %d", len(projects))
		}
	})
}

func TestDeleteProject(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	t.Run("existing project", func(t *testing.T) {
		req := models.VirtualDigSaveRequest{
			ProjectName: "Test",
			TerrainMap:  terrain,
		}
		project, _ := builder.SaveProject(req)

		ok := builder.DeleteProject(project.ID)
		if !ok {
			t.Error("Expected delete to succeed")
		}

		_, exists := builder.GetProject(project.ID)
		if exists {
			t.Error("Expected project to be deleted")
		}
	})

	t.Run("non existing project", func(t *testing.T) {
		ok := builder.DeleteProject("non-existent")
		if ok {
			t.Error("Expected delete to fail for non-existent project")
		}
	})
}

func TestFeasibilityEvaluation_Hydraulics(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	t.Run("no main channel", func(t *testing.T) {
		req := models.VirtualDigSaveRequest{
			ProjectName: "Test",
			TerrainMap:  terrain,
			Channels: []models.DigChannel{
				{
					IsMain: false,
					Points: []models.GeoPoint{
						{X: 1.0, Y: 0.5, Elevation: 845},
						{X: 1.0, Y: 4.0, Elevation: 820},
					},
				},
			},
		}
		project, _ := builder.SaveProject(req)
		if project.Feasibility.HydraulicScore >= 70 {
			t.Errorf("Expected low hydraulic score without main channel, got %f", project.Feasibility.HydraulicScore)
		}
	})

	t.Run("zero slope", func(t *testing.T) {
		req := models.VirtualDigSaveRequest{
			ProjectName: "Test",
			TerrainMap:  terrain,
			Channels: []models.DigChannel{
				{
					IsMain: true,
					Points: []models.GeoPoint{
						{X: 1.0, Y: 0.5, Elevation: 800},
						{X: 1.0, Y: 4.0, Elevation: 800},
					},
				},
			},
		}
		project, _ := builder.SaveProject(req)
		if project.Feasibility.HydraulicScore >= 60 {
			t.Errorf("Expected lower hydraulic score with zero slope, got %f", project.Feasibility.HydraulicScore)
		}
	})
}

func TestFeasibilityEvaluation_Geology(t *testing.T) {
	builder := New()

	t.Run("good soil gravel", func(t *testing.T) {
		terrain := builder.GetDefaultTerrain()
		terrain.SoilType = "gravel"

		req := models.VirtualDigSaveRequest{
			ProjectName: "Test",
			TerrainMap:  terrain,
			Channels: []models.DigChannel{
				{
					IsMain: true,
					Points: []models.GeoPoint{
						{X: 0.5, Y: 0.5, Elevation: 845},
						{X: 0.5, Y: 7.5, Elevation: 800},
					},
				},
			},
		}
		project, _ := builder.SaveProject(req)
		if project.Feasibility.GeologicalScore < 70 {
			t.Errorf("Expected good geological score with gravel, got %f", project.Feasibility.GeologicalScore)
		}
	})

	t.Run("poor soil clay", func(t *testing.T) {
		terrain := builder.GetDefaultTerrain()
		terrain.SoilType = "clay"

		req := models.VirtualDigSaveRequest{
			ProjectName: "Test",
			TerrainMap:  terrain,
			Channels: []models.DigChannel{
				{
					IsMain: true,
					Points: []models.GeoPoint{
						{X: 0.5, Y: 0.5, Elevation: 845},
						{X: 0.5, Y: 7.5, Elevation: 800},
					},
				},
			},
		}
		project, _ := builder.SaveProject(req)
		if project.Feasibility.GeologicalScore >= 70 {
			t.Errorf("Expected lower geological score with clay, got %f", project.Feasibility.GeologicalScore)
		}
	})
}

func TestCalculateChannelMetrics(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	t.Run("valid channel", func(t *testing.T) {
		ch := models.DigChannel{
			IsMain: true,
			Width:  1.5,
			Height: 2.0,
			Points: []models.GeoPoint{
				{X: 1.0, Y: 0.0, Elevation: 850},
				{X: 1.0, Y: 3.0, Elevation: 830},
			},
		}
		builder.calculateChannelMetrics(&ch, terrain)

		if ch.Length <= 0 {
			t.Error("Expected positive channel length")
		}
		if ch.Slope <= 0 {
			t.Error("Expected positive slope")
		}
	})

	t.Run("single point", func(t *testing.T) {
		ch := models.DigChannel{
			Points: []models.GeoPoint{{X: 1.0, Y: 1.0, Elevation: 800}},
		}
		builder.calculateChannelMetrics(&ch, terrain)

		if ch.Length != 0 {
			t.Errorf("Expected zero length for single point, got %f", ch.Length)
		}
		if ch.Slope != 0 {
			t.Errorf("Expected zero slope for single point, got %f", ch.Slope)
		}
	})

	t.Run("auto fill dimensions", func(t *testing.T) {
		ch := models.DigChannel{
			Points: []models.GeoPoint{
				{X: 1.0, Y: 0.0, Elevation: 850},
				{X: 1.0, Y: 3.0, Elevation: 830},
			},
		}
		builder.calculateChannelMetrics(&ch, terrain)

		if ch.Width == 0 {
			t.Error("Expected width to be auto-filled")
		}
		if ch.Height == 0 {
			t.Error("Expected height to be auto-filled")
		}
	})
}

func TestCalculateShaftMetrics(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	t.Run("reaches water", func(t *testing.T) {
		sh := models.DigShaft{
			Position: models.GeoPoint{X: 2.0, Y: 4.0, Elevation: 825},
			Depth:    60,
		}
		builder.calculateShaftMetrics(&sh, terrain)

		if !sh.ReachesWater {
			t.Error("Expected deep shaft to reach water")
		}
	})

	t.Run("does not reach water", func(t *testing.T) {
		sh := models.DigShaft{
			Position: models.GeoPoint{X: 2.0, Y: 4.0, Elevation: 825},
			Depth:    5,
		}
		builder.calculateShaftMetrics(&sh, terrain)

		if sh.ReachesWater {
			t.Error("Expected shallow shaft not to reach water")
		}
	})

	t.Run("auto fill", func(t *testing.T) {
		sh := models.DigShaft{
			Position: models.GeoPoint{X: 2.0, Y: 4.0, Elevation: 825},
		}
		builder.calculateShaftMetrics(&sh, terrain)

		if sh.Depth == 0 {
			t.Error("Expected depth to be auto-filled")
		}
		if sh.Diameter == 0 {
			t.Error("Expected diameter to be auto-filled")
		}
	})
}

func TestGetDigGuide(t *testing.T) {
	builder := New()
	guide := builder.GetDigGuide()

	if guide.GuideID == "" {
		t.Error("Expected guide ID")
	}

	if guide.GuideName == "" {
		t.Error("Expected guide name")
	}

	if len(guide.Steps) == 0 {
		t.Error("Expected non-empty steps")
	}

	if guide.CurrentStep < 1 {
		t.Error("Expected current step to be at least 1")
	}

	for _, step := range guide.Steps {
		if step.Title == "" {
			t.Error("Step title should not be empty")
		}
		if step.Description == "" {
			t.Error("Step description should not be empty")
		}
		if len(step.Tips) == 0 {
			t.Error("Step should have tips")
		}
	}
}

func TestGetDesignTemplates(t *testing.T) {
	builder := New()
	templates := builder.GetDesignTemplates()

	if len(templates) == 0 {
		t.Fatal("Expected non-empty templates")
	}

	expectedCount := 3
	if len(templates) != expectedCount {
		t.Errorf("Expected %d templates, got %d", expectedCount, len(templates))
	}

	for _, tpl := range templates {
		if tpl.TemplateID == "" {
			t.Error("Template ID should not be empty")
		}
		if tpl.TemplateName == "" {
			t.Error("Template name should not be empty")
		}
		if tpl.Difficulty == "" {
			t.Error("Template difficulty should not be empty")
		}
		if len(tpl.Channels) == 0 {
			t.Error("Template should have channels")
		}
	}
}

func TestGetQuickTips(t *testing.T) {
	builder := New()
	tips := builder.GetQuickTips()

	if len(tips) == 0 {
		t.Fatal("Expected non-empty tips")
	}

	expectedCount := 8
	if len(tips) != expectedCount {
		t.Errorf("Expected %d tips, got %d", expectedCount, len(tips))
	}

	for _, tip := range tips {
		if tip.TipID == "" {
			t.Error("Tip ID should not be empty")
		}
		if tip.Category == "" {
			t.Error("Tip category should not be empty")
		}
		if tip.Content == "" {
			t.Error("Tip content should not be empty")
		}
	}
}

func TestGetSoilFlowFactor(t *testing.T) {
	builder := New()

	testCases := []struct {
		soilType string
		expected float64
	}{
		{"gravel", 1.2},
		{"sand", 1.0},
		{"silt", 0.7},
		{"clay", 0.4},
		{"limestone", 0.9},
		{"unknown", 1.0},
	}

	for _, tc := range testCases {
		t.Run(tc.soilType, func(t *testing.T) {
			factor := builder.getSoilFlowFactor(tc.soilType)
			if factor != tc.expected {
				t.Errorf("Expected factor %f for %s, got %f", tc.expected, tc.soilType, factor)
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				req := models.VirtualDigSaveRequest{
					ProjectName: "Concurrent",
					TerrainMap:  terrain,
				}
				project, _ := builder.SaveProject(req)
				builder.GetProject(project.ID)
				builder.ListProjects()
				builder.DeleteProject(project.ID)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestEstimateFlow_DifferentSoils(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	soilTypes := []string{"gravel", "sand", "silt", "clay", "limestone"}
	flows := make(map[string]float64)

	for _, soilType := range soilTypes {
		t := terrain
		t.SoilType = soilType

		req := models.VirtualDigSaveRequest{
			ProjectName: "Test",
			TerrainMap:  t,
			Channels: []models.DigChannel{
				{
					IsMain: true,
					Width:  1.5,
					Height: 2.0,
					Points: []models.GeoPoint{
						{X: 2.0, Y: 0.5, Elevation: 845},
						{X: 2.0, Y: 7.5, Elevation: 805},
					},
				},
			},
			Shafts: []models.DigShaft{
				{Position: models.GeoPoint{X: 2.0, Y: 4.0, Elevation: 825}, Depth: 50},
			},
		}

		project, _ := builder.SaveProject(req)
		flows[soilType] = project.SimulatedFlow
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

func TestSaveProject_Issues(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	t.Run("insufficient shafts", func(t *testing.T) {
		req := models.VirtualDigSaveRequest{
			ProjectName: "Test",
			TerrainMap:  terrain,
			Channels: []models.DigChannel{
				{
					IsMain: true,
					Points: []models.GeoPoint{
						{X: 1.0, Y: 0.5, Elevation: 845},
						{X: 1.0, Y: 7.5, Elevation: 800},
					},
				},
			},
			Shafts: []models.DigShaft{
				{Position: models.GeoPoint{X: 1.0, Y: 4.0, Elevation: 820}, Depth: 10},
			},
		}
		project, _ := builder.SaveProject(req)

		hasInsufficientShafts := false
		for _, issue := range project.Feasibility.Issues {
			if issue.Severity == "警告" {
				hasInsufficientShafts = true
				break
			}
		}
		if !hasInsufficientShafts {
			t.Error("Expected warning about insufficient shafts")
		}
	})
}

func TestCalculateStatistics(t *testing.T) {
	builder := New()

	channels := []models.DigChannel{
		{Length: 1000, Width: 1.5, Height: 2.0, Depth: 30},
		{Length: 500, Width: 1.0, Height: 1.5, Depth: 25},
	}
	shafts := []models.DigShaft{
		{Depth: 40, Diameter: 1.2},
		{Depth: 35, Diameter: 1.2},
	}

	stats := builder.calculateStatistics(channels, shafts)

	if stats.TotalChannelLength != 1500 {
		t.Errorf("Expected total length 1500, got %f", stats.TotalChannelLength)
	}

	if stats.TotalShafts != 2 {
		t.Errorf("Expected 2 shafts, got %d", stats.TotalShafts)
	}

	if stats.TotalExcavationVolume <= 0 {
		t.Error("Expected positive excavation volume")
	}

	if stats.EstimatedCost <= 0 {
		t.Error("Expected positive estimated cost")
	}
}

func TestFeasibility_CriticalIssueVeto(t *testing.T) {
	builder := New()
	terrain := builder.GetDefaultTerrain()

	req := models.VirtualDigSaveRequest{
		ProjectName: "Critical Test",
		TerrainMap:  terrain,
		Channels: []models.DigChannel{
			{
				IsMain: true,
				Points: []models.GeoPoint{
					{X: 2.0, Y: 0.5, Elevation: 800},
					{X: 2.0, Y: 7.5, Elevation: 810},
				},
			},
		},
		Shafts: []models.DigShaft{
			{Position: models.GeoPoint{X: 2.0, Y: 1.0, Elevation: 800}, Depth: 5},
			{Position: models.GeoPoint{X: 2.0, Y: 4.0, Elevation: 800}, Depth: 5},
			{Position: models.GeoPoint{X: 2.0, Y: 7.0, Elevation: 800}, Depth: 5},
		},
	}

	project, _ := builder.SaveProject(req)

	if project.Feasibility.IsFeasible {
		t.Error("Project with critical issues should not be feasible")
	}

	hasCritical := false
	for _, issue := range project.Feasibility.Issues {
		if issue.Severity == "严重" {
			hasCritical = true
			break
		}
	}
	if !hasCritical {
		t.Error("Expected at least one critical issue")
	}
}
