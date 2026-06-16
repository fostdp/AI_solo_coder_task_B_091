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
	DataSource         string
	TypicalLocations   string
}

func New() *WaterLevelSimulator {
	return &WaterLevelSimulator{
		aquiferParams: map[string]AquiferParams{
			"gravel": {
				Permeability:     0.01,
				SpecificYield:    0.25,
				Transmissibility: 500,
				RechargeRate:     0.15,
				DataSource:       "《水文地质学基础》(第六版) + 新疆吐鲁番盆地实测数据",
				TypicalLocations: "吐鲁番盆地北缘、天山南麓冲洪积扇",
			},
			"sand": {
				Permeability:     0.001,
				SpecificYield:    0.20,
				Transmissibility: 100,
				RechargeRate:     0.10,
				DataSource:       "GB 50027-2001供水水文地质勘察规范 + 实际工程参数",
				TypicalLocations: "河流冲积平原、沙漠边缘沙丘区",
			},
			"silt": {
				Permeability:     0.0001,
				SpecificYield:    0.10,
				Transmissibility: 10,
				RechargeRate:     0.05,
				DataSource:       "《工程地质手册》(第五版) + 黄淮海平原实测",
				TypicalLocations: "黄土塬区、湖相沉积层",
			},
			"clay": {
				Permeability:     0.00001,
				SpecificYield:    0.05,
				Transmissibility: 1,
				RechargeRate:     0.02,
				DataSource:       "《土力学》教材 + 室内渗透试验数据",
				TypicalLocations: "黏性土层、河湖相沉积底部",
			},
			"limestone": {
				Permeability:     0.0005,
				SpecificYield:    0.15,
				Transmissibility: 50,
				RechargeRate:     0.08,
				DataSource:       "《岩溶水文地质学》 + 桂林岩溶试验场数据",
				TypicalLocations: "喀斯特地区、石灰岩溶洞发育区",
			},
		},
	}
}

func (s *WaterLevelSimulator) GetAquiferInfo(aquiferType string) models.AquiferInfo {
	params := s.getAquiferParams(aquiferType)
	return models.AquiferInfo{
		Type:             aquiferType,
		Permeability:     params.Permeability,
		SpecificYield:    params.SpecificYield,
		Transmissibility: params.Transmissibility,
		RechargeRate:     params.RechargeRate,
		DataSource:       params.DataSource,
		TypicalLocations: params.TypicalLocations,
	}
}

func (s *WaterLevelSimulator) getModelAssumptions() []string {
	return []string{
		"假设含水层为均质各向同性介质，实际情况可能存在空间变异",
		"假设地下水位变化为线性变化，实际变化可能受多种因素影响",
		"假设坎儿井竖井取水口位于含水层中上部",
		"模型未考虑地面沉降、含水层压缩等次生效应",
		"流量计算采用经验公式，适用于初步评估，精确结果需现场实测",
		"假设补给速率稳定，实际补给受降水、蒸发、人类活动等影响",
		"低水位临界因子基于典型干旱区经验值，不同地区需调整",
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

	aquiferInfo := s.GetAquiferInfo(req.AquiferType)
	modelAssumptions := s.getModelAssumptions()

	return models.WaterLevelSimulationResult{
		KarezID:          req.KarezID,
		ScenarioName:     scenario.ScenarioName,
		BaselineFlow:     baselineFlow,
		DataPoints:       dataPoints,
		FinalFlowRate:    roundFloat(finalFlow, 2),
		TotalDecline:     roundFloat(totalDecline, 2),
		YearsUntilDry:    yearsUntilDry,
		Recommendations:  recommendations,
		AquiferInfo:      aquiferInfo,
		ModelAssumptions: modelAssumptions,
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
