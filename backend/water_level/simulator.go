package waterlevel

import (
	"math"

	"karez-system/models"
)

type WaterLevelSimulator struct {
	aquiferParams map[string]AquiferParams
}

type AquiferParams struct {
	Permeability       float64
	SpecificYield      float64
	Transmissibility   float64
	RechargeRate       float64
}

func New() *WaterLevelSimulator {
	return &WaterLevelSimulator{
		aquiferParams: map[string]AquiferParams{
			"gravel": {
				Permeability:     0.01,
				SpecificYield:    0.25,
				Transmissibility: 500,
				RechargeRate:     0.15,
			},
			"sand": {
				Permeability:     0.001,
				SpecificYield:    0.20,
				Transmissibility: 100,
				RechargeRate:     0.10,
			},
			"silt": {
				Permeability:     0.0001,
				SpecificYield:    0.10,
				Transmissibility: 10,
				RechargeRate:     0.05,
			},
			"clay": {
				Permeability:     0.00001,
				SpecificYield:    0.05,
				Transmissibility: 1,
				RechargeRate:     0.02,
			},
			"limestone": {
				Permeability:     0.0005,
				SpecificYield:    0.15,
				Transmissibility: 50,
				RechargeRate:     0.08,
			},
		},
	}
}

func (s *WaterLevelSimulator) getAquiferParams(aquiferType string) AquiferParams {
	if params, exists := s.aquiferParams[aquiferType]; exists {
		return params
	}
	return s.aquiferParams["gravel"]
}

func (s *WaterLevelSimulator) GetDefaultScenarios() []models.WaterLevelScenario {
	return []models.WaterLevelScenario{
		{
			ScenarioName:      "稳定状态",
			InitialWaterLevel: 30,
			TargetWaterLevel:  30,
			ChangeRate:        0,
			DurationYears:     50,
			Description:       "地下水位保持不变，模拟自然平衡状态下的坎儿井出水量。",
		},
		{
			ScenarioName:      "缓慢下降",
			InitialWaterLevel: 30,
			TargetWaterLevel:  20,
			ChangeRate:        0.2,
			DurationYears:     50,
			Description:       "每年地下水位下降0.2米，模拟气候变化影响下的缓慢衰退。",
		},
		{
			ScenarioName:      "中度下降",
			InitialWaterLevel: 30,
			TargetWaterLevel:  10,
			ChangeRate:        0.5,
			DurationYears:     40,
			Description:       "每年地下水位下降0.5米，模拟农业过度用水导致的水位下降。",
		},
		{
			ScenarioName:      "急剧下降",
			InitialWaterLevel: 30,
			TargetWaterLevel:  5,
			ChangeRate:        1.0,
			DurationYears:     25,
			Description:       "每年地下水位下降1.0米，模拟大量机井抽水导致的严重水位下降。",
		},
		{
			ScenarioName:      "生态恢复",
			InitialWaterLevel: 15,
			TargetWaterLevel:  30,
			ChangeRate:        -0.3,
			DurationYears:     50,
			Description:       "通过节水和生态补水，地下水位每年回升0.3米，模拟生态修复效果。",
		},
	}
}

func (s *WaterLevelSimulator) SimulateWaterLevelImpact(req models.WaterLevelSimulationRequest) []models.WaterLevelSimulationResult {
	if len(req.Scenarios) == 0 {
		req.Scenarios = s.GetDefaultScenarios()
	}
	if req.BaselineFlowRate == 0 {
		req.BaselineFlowRate = 3000
	}
	if req.ShaftDepth == 0 {
		req.ShaftDepth = 40
	}
	if req.AquiferType == "" {
		req.AquiferType = "gravel"
	}

	results := make([]models.WaterLevelSimulationResult, 0, len(req.Scenarios))
	for _, scenario := range req.Scenarios {
		result := s.simulateSingleScenario(req, scenario)
		results = append(results, result)
	}
	return results
}

func (s *WaterLevelSimulator) simulateSingleScenario(
	req models.WaterLevelSimulationRequest,
	scenario models.WaterLevelScenario,
) models.WaterLevelSimulationResult {
	aquiferParams := s.getAquiferParams(req.AquiferType)
	dataPoints := make([]models.WaterLevelDataPoint, 0)
	baselineFlow := req.BaselineFlowRate
	shaftDepth := req.ShaftDepth
	yearsUntilDry := -1

	currentWaterLevel := scenario.InitialWaterLevel
	totalDuration := scenario.DurationYears
	if totalDuration == 0 {
		levelDiff := math.Abs(scenario.TargetWaterLevel - scenario.InitialWaterLevel)
		if scenario.ChangeRate > 0 {
			totalDuration = int(math.Ceil(levelDiff / scenario.ChangeRate))
		} else {
			totalDuration = 50
		}
	}
	if totalDuration > 100 {
		totalDuration = 100
	}

	for year := 0; year <= totalDuration; year++ {
		if scenario.ChangeRate != 0 {
			progress := float64(year) / float64(totalDuration)
			if progress > 1 {
				progress = 1
			}
			currentWaterLevel = scenario.InitialWaterLevel +
				(scenario.TargetWaterLevel-scenario.InitialWaterLevel)*progress
		}

		shaftIntakeDepth := shaftDepth - currentWaterLevel
		if shaftIntakeDepth < 0 {
			shaftIntakeDepth = 0
		}

		flowRate := s.calculateFlowRate(
			baselineFlow,
			currentWaterLevel,
			shaftDepth,
			shaftIntakeDepth,
			aquiferParams,
		)

		flowChangePercent := 0.0
		if baselineFlow > 0 {
			flowChangePercent = (flowRate - baselineFlow) / baselineFlow * 100
		} else {
			flowChangePercent = 0
		}

		isSustained := flowRate > 0.001

		warningLevel := "正常"
		if baselineFlow > 0 {
			switch {
			case flowRate < baselineFlow*0.1:
				warningLevel = "危急"
			case flowRate < baselineFlow*0.3:
				warningLevel = "严重"
			case flowRate < baselineFlow*0.5:
				warningLevel = "警告"
			case flowRate < baselineFlow*0.8:
				warningLevel = "注意"
			}
		} else {
			if flowRate <= 0 {
				warningLevel = "危急"
			}
		}

		if yearsUntilDry < 0 && !isSustained {
			yearsUntilDry = year
		}

		dataPoints = append(dataPoints, models.WaterLevelDataPoint{
			Year:              year,
			WaterLevel:        roundFloat(currentWaterLevel, 2),
			FlowRate:          roundFloat(flowRate, 2),
			FlowChangePercent: roundFloat(flowChangePercent, 2),
			ShaftIntakeDepth:  roundFloat(shaftIntakeDepth, 2),
			IsFlowSustained:   isSustained,
			WarningLevel:      warningLevel,
		})
	}

	finalFlow := dataPoints[len(dataPoints)-1].FlowRate
	totalDecline := 0.0
	if baselineFlow > 0 {
		totalDecline = (baselineFlow - finalFlow) / baselineFlow * 100
	}
	if totalDecline < 0 && scenario.ChangeRate >= 0 {
		totalDecline = 0
	}

	recommendations := s.generateRecommendations(
		scenario,
		finalFlow,
		baselineFlow,
		yearsUntilDry,
	)

	return models.WaterLevelSimulationResult{
		KarezID:         req.KarezID,
		ScenarioName:    scenario.ScenarioName,
		BaselineFlow:    baselineFlow,
		DataPoints:      dataPoints,
		FinalFlowRate:   roundFloat(finalFlow, 2),
		TotalDecline:    roundFloat(totalDecline, 2),
		YearsUntilDry:   yearsUntilDry,
		Recommendations: recommendations,
	}
}

func (s *WaterLevelSimulator) calculateFlowRate(
	baselineFlow, waterLevel, shaftDepth, shaftIntakeDepth float64,
	aquiferParams AquiferParams,
) float64 {
	if waterLevel <= 0 {
		return 0
	}

	effectiveDepth := waterLevel
	if shaftIntakeDepth > 0 {
		effectiveDepth = waterLevel + shaftIntakeDepth*0.3
	}

	depthFactor := 0.0
	if shaftDepth > 0 {
		depthRatio := effectiveDepth / shaftDepth
		if depthRatio > 1 {
			depthRatio = 1
		}
		depthFactor = math.Pow(depthRatio, 0.05)
	}

	permeabilityFactor := aquiferParams.Permeability / 0.01
	if permeabilityFactor > 1.1 {
		permeabilityFactor = 1.1
	}
	if permeabilityFactor < 0.7 {
		permeabilityFactor = 0.7
	}

	rechargeFactor := 1.0 + aquiferParams.RechargeRate*0.02

	flowRate := baselineFlow * depthFactor * permeabilityFactor * rechargeFactor

	if waterLevel < 10 {
		criticalFactor := waterLevel / 10
		flowRate *= criticalFactor * criticalFactor
	}

	if waterLevel <= 5 {
		criticalFactor := waterLevel / 5
		if waterLevel <= 0 {
			criticalFactor = 0
		}
		flowRate *= criticalFactor * criticalFactor * criticalFactor * criticalFactor
	}

	return math.Max(0, flowRate)
}

func (s *WaterLevelSimulator) generateRecommendations(
	scenario models.WaterLevelScenario,
	finalFlow, baselineFlow float64,
	yearsUntilDry int,
) []string {
	recommendations := make([]string, 0)

	if scenario.ChangeRate > 0.8 {
		recommendations = append(recommendations,
			"立即关停周边不必要的机井，控制地下水开采强度。")
		recommendations = append(recommendations,
			"实施紧急生态补水工程，从附近河流或水库向地下含水层回灌。")
	} else if scenario.ChangeRate > 0.3 {
		recommendations = append(recommendations,
			"限制农业灌溉机井的抽水量，推广节水灌溉技术。")
		recommendations = append(recommendations,
			"建立地下水水位监测预警系统，实时跟踪变化趋势。")
	}

	if yearsUntilDry > 0 && yearsUntilDry < 30 {
		recommendations = append(recommendations,
			"预计约"+intToStr(yearsUntilDry)+"年后坎儿井可能干涸，建议提前制定替代供水方案。")
	}

	if baselineFlow > 0 && finalFlow/baselineFlow < 0.3 {
		recommendations = append(recommendations,
			"对现有坎儿井进行深度清淤和防渗加固，尽可能恢复集水能力。")
		recommendations = append(recommendations,
			"考虑在坎儿井集水区修建人工渗水池，增加地下水补给。")
	}

	if scenario.ChangeRate < 0 {
		recommendations = append(recommendations,
			"生态恢复方案效果良好，建议持续坚持节水和补水措施。")
		recommendations = append(recommendations,
			"在水位恢复过程中，定期检查坎儿井结构安全性，防止塌陷。")
	}

	recommendations = append(recommendations,
		"发展坎儿井文化旅游业，以旅游收入反哺坎儿井保护和维护。")
	recommendations = append(recommendations,
		"建立社区参与的坎儿井管理委员会，形成长效保护机制。")

	return recommendations
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
