package virtualdig

import (
	"crypto/rand"
	"encoding/hex"
	"math"
	"sync"
	"time"

	"karez-system/models"
)

type VirtualDigService struct {
	mu       sync.RWMutex
	projects map[string]*models.VirtualDigProject
}

func New() *VirtualDigService {
	return &VirtualDigService{
		projects: make(map[string]*models.VirtualDigProject),
	}
}

func (s *VirtualDigService) generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *VirtualDigService) GetDefaultTerrain() models.TerrainConfig {
	return models.TerrainConfig{
		WidthKm:         5.0,
		LengthKm:        8.0,
		HeadElevation:   850,
		TailElevation:   800,
		WaterTableDepth: 30,
		SoilType:        "gravel",
		Obstacles: []models.MapObstacle{
			{
				ID:     "obs1",
				Type:   "rock",
				X:      2.5,
				Y:      3.0,
				Radius: 0.3,
				Label:  "花岗岩体",
			},
			{
				ID:     "obs2",
				Type:   "clay",
				X:      4.0,
				Y:      5.5,
				Radius: 0.4,
				Label:  "黏土层",
			},
			{
				ID:     "obs3",
				Type:   "fault",
				X:      1.5,
				Y:      6.0,
				Radius: 0.25,
				Label:  "地质断层",
			},
		},
	}
}

func (s *VirtualDigService) SaveProject(req models.VirtualDigSaveRequest) (*models.VirtualDigProject, error) {
	project := &models.VirtualDigProject{
		ID:          s.generateID(),
		ProjectName: req.ProjectName,
		Creator:     req.Creator,
		CreatedAt:   time.Now(),
		TerrainMap:  req.TerrainMap,
		Channels:    req.Channels,
		Shafts:      req.Shafts,
	}

	for i := range project.Channels {
		if project.Channels[i].ID == "" {
			project.Channels[i].ID = "ch_" + s.generateID()
		}
		if project.Channels[i].Name == "" {
			project.Channels[i].Name = "暗渠" + intToStr(i+1)
		}
		s.calculateChannelMetrics(&project.Channels[i], project.TerrainMap)
	}

	for i := range project.Shafts {
		if project.Shafts[i].ID == "" {
			project.Shafts[i].ID = "sh_" + s.generateID()
		}
		if project.Shafts[i].Name == "" {
			project.Shafts[i].Name = "竖井" + intToStr(i+1)
		}
		s.calculateShaftMetrics(&project.Shafts[i], project.TerrainMap)
	}

	project.Statistics = s.calculateStatistics(project.Channels, project.Shafts)
	project.SimulatedFlow = s.estimateFlow(project.Channels, project.Shafts, project.TerrainMap)
	project.Feasibility = s.evaluateFeasibility(project)

	s.mu.Lock()
	s.projects[project.ID] = project
	s.mu.Unlock()

	return project, nil
}

func (s *VirtualDigService) SimulateDesign(req models.VirtualDigSimulateRequest) (*models.VirtualDigProject, error) {
	return s.SaveProject(models.VirtualDigSaveRequest{
		ProjectName: "模拟项目_" + s.generateID(),
		Creator:     "system",
		TerrainMap:  req.TerrainMap,
		Channels:    req.Channels,
		Shafts:      req.Shafts,
	})
}

func (s *VirtualDigService) GetProject(id string) (*models.VirtualDigProject, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	project, exists := s.projects[id]
	return project, exists
}

func (s *VirtualDigService) ListProjects() []*models.VirtualDigProject {
	s.mu.RLock()
	defer s.mu.RUnlock()
	projects := make([]*models.VirtualDigProject, 0, len(s.projects))
	for _, p := range s.projects {
		projects = append(projects, p)
	}
	return projects
}

func (s *VirtualDigService) DeleteProject(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.projects[id]; exists {
		delete(s.projects, id)
		return true
	}
	return false
}

func (s *VirtualDigService) calculateChannelMetrics(ch *models.DigChannel, terrain models.TerrainConfig) {
	if len(ch.Points) < 2 {
		ch.Length = 0
		ch.Slope = 0
		return
	}

	totalLength := 0.0
	for i := 1; i < len(ch.Points); i++ {
		dx := ch.Points[i].X - ch.Points[i-1].X
		dy := ch.Points[i].Y - ch.Points[i-1].Y
		dist := math.Sqrt(dx*dx + dy*dy)
		totalLength += dist
	}
	ch.Length = totalLength * 1000

	headElev := ch.Points[0].Elevation
	tailElev := ch.Points[len(ch.Points)-1].Elevation
	if ch.Length > 0 {
		ch.Slope = (headElev - tailElev) / ch.Length
	}

	if ch.Width == 0 {
		ch.Width = 1.2
	}
	if ch.Height == 0 {
		ch.Height = 1.8
	}
	if ch.Depth == 0 {
		avgY := 0.0
		for _, p := range ch.Points {
			avgY += p.Y
		}
		avgY /= float64(len(ch.Points))
		terrainElev := terrain.HeadElevation - (terrain.HeadElevation-terrain.TailElevation)*avgY/terrain.LengthKm
		ch.Depth = terrainElev - headElev + terrain.WaterTableDepth*0.5
		if ch.Depth < 10 {
			ch.Depth = 10
		}
	}
}

func (s *VirtualDigService) calculateShaftMetrics(sh *models.DigShaft, terrain models.TerrainConfig) {
	terrainElev := terrain.HeadElevation -
		(terrain.HeadElevation-terrain.TailElevation)*sh.Position.Y/terrain.LengthKm

	if sh.Depth == 0 {
		sh.Depth = terrainElev - sh.Position.Elevation + terrain.WaterTableDepth
		if sh.Depth < 15 {
			sh.Depth = 15
		}
	}

	if sh.Diameter == 0 {
		sh.Diameter = 1.2
	}

	shaftBottomElev := sh.Position.Elevation - sh.Depth
	waterTableElev := terrainElev - terrain.WaterTableDepth
	sh.ReachesWater = shaftBottomElev <= waterTableElev

	if sh.DistanceFromHead == 0 {
		sh.DistanceFromHead = sh.Position.Y * 1000
	}
}

func (s *VirtualDigService) calculateStatistics(
	channels []models.DigChannel,
	shafts []models.DigShaft,
) models.DigStatistics {
	totalLength := 0.0
	totalVolume := 0.0
	totalDepth := 0.0

	for _, ch := range channels {
		totalLength += ch.Length
		channelVolume := ch.Length * ch.Width * ch.Height
		totalVolume += channelVolume
		totalDepth += ch.Depth
	}

	for _, sh := range shafts {
		shaftVolume := math.Pi * (sh.Diameter/2) * (sh.Diameter/2) * sh.Depth
		totalVolume += shaftVolume
		totalDepth += sh.Depth
	}

	totalItems := float64(len(channels) + len(shafts))
	avgDepth := 0.0
	if totalItems > 0 {
		avgDepth = totalDepth / totalItems
	}

	manDaysPerM3 := 0.8
	estimatedManDays := totalVolume * manDaysPerM3
	costPerManDay := 300.0
	estimatedCost := estimatedManDays * costPerManDay

	return models.DigStatistics{
		TotalChannelLength:   roundFloat(totalLength, 2),
		TotalShafts:          len(shafts),
		TotalExcavationVolume: roundFloat(totalVolume, 2),
		EstimatedManDays:     roundFloat(estimatedManDays, 2),
		EstimatedCost:        roundFloat(estimatedCost, 2),
		AverageDepth:         roundFloat(avgDepth, 2),
	}
}

func (s *VirtualDigService) estimateFlow(
	channels []models.DigChannel,
	shafts []models.DigShaft,
	terrain models.TerrainConfig,
) float64 {
	if len(channels) == 0 {
		return 0
	}

	totalFlow := 0.0
	waterReachingShafts := 0
	for _, sh := range shafts {
		if sh.ReachesWater {
			waterReachingShafts++
		}
	}

	for _, ch := range channels {
		if !ch.IsMain {
			continue
		}

		width := ch.Width
		height := ch.Height
		slope := ch.Slope

		if slope <= 0 {
			slope = 0.002
		}
		if slope > 0.02 {
			slope = 0.02
		}

		waterDepth := height * 0.6
		area := width * waterDepth
		wettedPerimeter := width + 2*waterDepth
		hydraulicRadius := area / wettedPerimeter
		if hydraulicRadius <= 0 {
			continue
		}

		roughness := 0.015
		velocity := (1.0 / roughness) * math.Pow(hydraulicRadius, 2.0/3.0) * math.Sqrt(slope)
		flowRate := area * velocity

		shaftFactor := 1.0
		if len(shafts) > 0 {
			shaftFactor = 0.5 + 0.5*float64(waterReachingShafts)/float64(len(shafts))
		}
		flowRate *= shaftFactor

		soilFactor := s.getSoilFlowFactor(terrain.SoilType)
		flowRate *= soilFactor

		totalFlow += flowRate
	}

	return roundFloat(totalFlow*86400, 2)
}

func (s *VirtualDigService) getSoilFlowFactor(soilType string) float64 {
	switch soilType {
	case "gravel":
		return 1.2
	case "sand":
		return 1.0
	case "silt":
		return 0.7
	case "clay":
		return 0.4
	case "limestone":
		return 0.9
	default:
		return 1.0
	}
}

func (s *VirtualDigService) evaluateFeasibility(project *models.VirtualDigProject) models.FeasibilityReport {
	issues := make([]models.FeasibilityIssue, 0)
	suggestions := make([]string, 0)

	hydraulicScore := s.evaluateHydraulics(project, &issues, &suggestions)
	geologicalScore := s.evaluateGeology(project, &issues, &suggestions)
	economicScore := s.evaluateEconomics(project, &issues, &suggestions)

	overallScore := (hydraulicScore + geologicalScore + economicScore) / 3.0

	hasCriticalIssue := false
	for _, issue := range issues {
		if issue.Severity == "严重" {
			hasCriticalIssue = true
			break
		}
	}

	isFeasible := overallScore >= 50 && !hasCriticalIssue

	if overallScore >= 80 {
		suggestions = append(suggestions, "设计方案优秀，可进入施工准备阶段。")
	} else if overallScore >= 60 {
		suggestions = append(suggestions, "设计方案基本可行，建议根据上述问题优化后实施。")
	} else if overallScore >= 40 {
		suggestions = append(suggestions, "设计方案存在较多问题，建议大幅调整后重新评估。")
	} else {
		suggestions = append(suggestions, "设计方案可行性较低，建议重新选址或改变设计思路。")
	}

	return models.FeasibilityReport{
		IsFeasible:      isFeasible,
		OverallScore:    roundFloat(overallScore, 1),
		HydraulicScore:  roundFloat(hydraulicScore, 1),
		GeologicalScore: roundFloat(geologicalScore, 1),
		EconomicScore:   roundFloat(economicScore, 1),
		Issues:          issues,
		Suggestions:     suggestions,
	}
}

func (s *VirtualDigService) evaluateHydraulics(
	project *models.VirtualDigProject,
	issues *[]models.FeasibilityIssue,
	suggestions *[]string,
) float64 {
	score := 70.0

	hasMainChannel := false
	for _, ch := range project.Channels {
		if ch.IsMain {
			hasMainChannel = true
			if ch.Slope <= 0 {
				score -= 20
				*issues = append(*issues, models.FeasibilityIssue{
					Severity: "严重",
					Message:  "主暗渠坡度小于等于零，水流无法自流。",
					Location: ch.Name,
				})
			} else if ch.Slope < 0.001 {
				score -= 10
				*issues = append(*issues, models.FeasibilityIssue{
					Severity: "警告",
					Message:  "主暗渠坡度过小，可能导致淤积。",
					Location: ch.Name,
				})
			} else if ch.Slope > 0.015 {
				score -= 5
				*issues = append(*issues, models.FeasibilityIssue{
					Severity: "注意",
					Message:  "主暗渠坡度过大，水流冲刷可能损坏渠道。",
					Location: ch.Name,
				})
			}

			if ch.Length < 500 {
				score -= 5
				*suggestions = append(*suggestions,
					ch.Name+"长度较短，建议延长以增加集水面积。")
			}
		}
	}

	if !hasMainChannel {
		score -= 30
		*issues = append(*issues, models.FeasibilityIssue{
			Severity: "严重",
			Message:  "缺少主暗渠，无法形成完整的输水系统。",
		})
	}

	if len(project.Shafts) < 3 && len(project.Channels) > 0 {
		score -= 25
		*issues = append(*issues, models.FeasibilityIssue{
			Severity: "警告",
			Message:  "竖井数量不足，不利于施工通风和后期维护。",
		})
	}

	waterReachingCount := 0
	for _, sh := range project.Shafts {
		if sh.ReachesWater {
			waterReachingCount++
		}
	}
	if len(project.Shafts) > 0 && float64(waterReachingCount) < float64(len(project.Shafts))/2 {
		score -= 35
		*issues = append(*issues, models.FeasibilityIssue{
			Severity: "严重",
			Message:  "超过半数竖井未触及地下水位，集水能力不足。",
		})
	}

	if score < 0 {
		score = 0
	}
	return score
}

func (s *VirtualDigService) evaluateGeology(
	project *models.VirtualDigProject,
	issues *[]models.FeasibilityIssue,
	suggestions *[]string,
) float64 {
	score := 75.0

	obstacleHits := 0
	for _, ch := range project.Channels {
		for _, pt := range ch.Points {
			for _, obs := range project.TerrainMap.Obstacles {
				dx := pt.X - obs.X
				dy := pt.Y - obs.Y
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist < obs.Radius {
					obstacleHits++
					severity := "警告"
					if obs.Type == "fault" {
						severity = "严重"
						score -= 15
					} else if obs.Type == "rock" {
						score -= 8
					} else {
						score -= 5
					}
					*issues = append(*issues, models.FeasibilityIssue{
						Severity: severity,
						Message:  "暗渠经过" + obs.Label + "区域，施工难度增加。",
						Location: ch.Name,
					})
				}
			}
		}
	}

	soilType := project.TerrainMap.SoilType
	switch soilType {
	case "gravel":
		score += 10
		*suggestions = append(*suggestions, "砾石地层渗透性好，集水效果佳。")
	case "sand":
		score += 5
	case "silt":
		score -= 5
		*suggestions = append(*suggestions, "粉土地层需注意渠道防渗处理。")
	case "clay":
		score -= 15
		*issues = append(*issues, models.FeasibilityIssue{
			Severity: "警告",
			Message:  "黏土地层渗透性差，集水效率低。",
		})
	case "limestone":
		score += 0
		*suggestions = append(*suggestions, "石灰岩地层可能存在溶洞，施工前需详细勘探。")
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

func (s *VirtualDigService) evaluateEconomics(
	project *models.VirtualDigProject,
	issues *[]models.FeasibilityIssue,
	suggestions *[]string,
) float64 {
	score := 70.0

	stats := project.Statistics
	flow := project.SimulatedFlow

	if flow > 0 {
		costPerM3 := stats.EstimatedCost / flow
		if costPerM3 < 50 {
			score += 20
			*suggestions = append(*suggestions, "单位水量成本低廉，经济效益良好。")
		} else if costPerM3 < 100 {
			score += 10
		} else if costPerM3 < 200 {
			score -= 5
		} else {
			score -= 20
			*issues = append(*issues, models.FeasibilityIssue{
				Severity: "警告",
				Message:  "单位水量建设成本过高，经济性不佳。",
			})
		}
	}

	if stats.TotalExcavationVolume > 100000 {
		score -= 10
		*issues = append(*issues, models.FeasibilityIssue{
			Severity: "注意",
			Message:  "工程量较大，建议分期分段施工。",
		})
	}

	if stats.EstimatedManDays > 50000 {
		score -= 10
		*issues = append(*issues, models.FeasibilityIssue{
			Severity: "注意",
			Message:  "人工需求量大，建议采用机械化施工降低人力成本。",
		})
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

func roundFloat(value float64, decimals int) float64 {
	shift := math.Pow(10, float64(decimals))
	return math.Round(value*shift) / shift
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	digits := []rune{}
	for n > 0 {
		digits = append([]rune{rune('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]rune{'-'}, digits...)
	}
	return string(digits)
}
