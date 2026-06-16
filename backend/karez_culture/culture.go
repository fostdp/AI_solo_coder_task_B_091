package karezculture

import (
	"karez-system/models"
)

type CultureService struct{}

func New() *CultureService {
	return &CultureService{}
}

func (s *CultureService) GetTechnologyEvolution() *models.TechnologyEvolutionAnalysis {
	evolutions := []models.EraTechnology{
		{
			Era:              "西汉时期",
			TimePeriod:       "公元前206年 - 公元8年",
			KeyFeatures:      []string{"初创阶段", "小型暗渠", "人工挖掘", "无衬砌"},
			Materials:        []string{"原土", "木桩支护"},
			ConstructionTools: []string{"铁锨", "镢头", "背篓", "辘轳"},
			AverageDepth:     15,
			AverageLength:    1.5,
			MaxFlowRate:      500,
			WaterLossRate:    40,
			LabourRequirement: 5000,
			MaintenanceCycle: "每年清淤一次",
			HistoricalNotes:  "张骞通西域后，从中亚传入坎儿井技术，初期主要用于小规模农田灌溉。",
		},
		{
			Era:              "魏晋南北朝",
			TimePeriod:       "公元220年 - 589年",
			KeyFeatures:      []string{"规模扩大", "多竖井布局", "木衬砌出现", "分支渠系形成"},
			Materials:        []string{"木材", "卵石", "黏土"},
			ConstructionTools: []string{"铁锤", "铁钎", "滑轮", "木制运水车"},
			AverageDepth:     25,
			AverageLength:    3.0,
			MaxFlowRate:      1500,
			WaterLossRate:    30,
			LabourRequirement: 8000,
			MaintenanceCycle: "每半年检查一次，每年清淤",
			HistoricalNotes:  "中原战乱导致人口西迁，促进了西域水利工程发展，坎儿井技术逐步成熟。",
		},
		{
			Era:              "隋唐时期",
			TimePeriod:       "公元581年 - 907年",
			KeyFeatures:      []string{"系统化设计", "砖石衬砌", "精确坡度控制", "龙头工程出现"},
			Materials:        []string{"青砖", "条石", "石灰砂浆", "木材"},
			ConstructionTools: []string{"精密水准仪（简易）", "石凿", "铁制工具组", "人力绞车"},
			AverageDepth:     35,
			AverageLength:    5.0,
			MaxFlowRate:      3000,
			WaterLossRate:    20,
			LabourRequirement: 12000,
			MaintenanceCycle: "季度巡检，年度大修",
			HistoricalNotes:  "唐代安西都护府设立专门水利机构，坎儿井建造技术达到古代高峰。",
		},
		{
			Era:              "宋元时期",
			TimePeriod:       "公元960年 - 1368年",
			KeyFeatures:      []string{"地下水库连接", "多条暗渠并联", "冰窖储水技术", "水文地质勘测进步"},
			Materials:        []string{"烧制红砖", "花岗岩", "桐油灰缝", "竹制管道"},
			ConstructionTools: []string{"罗盘定位仪", "深层钻探工具", "炸药雏形", "水力提升机"},
			AverageDepth:     45,
			AverageLength:    8.0,
			MaxFlowRate:      5000,
			WaterLossRate:    15,
			LabourRequirement: 18000,
			MaintenanceCycle: "月度巡检，季度维护，年度大修",
			HistoricalNotes:  "西辽和元朝时期，中亚工匠带来更先进的地下工程技术，坎儿井体系进一步完善。",
		},
		{
			Era:              "明清时期",
			TimePeriod:       "公元1368年 - 1912年",
			KeyFeatures:      []string{"官督民办", "标准化施工", "完整水系网络", "水磨联动"},
			Materials:        []string{"窑烧青砖", "规格条石", "糯米灰浆", "铸铁构件"},
			ConstructionTools: []string{"精确测斜仪", "蒸汽抽水机（晚清）", "钢制工具", "轨道运土车"},
			AverageDepth:     60,
			AverageLength:    12.0,
			MaxFlowRate:      8000,
			WaterLossRate:    10,
			LabourRequirement: 25000,
			MaintenanceCycle: "专业水利营管理，定期维护制度完善",
			HistoricalNotes:  "清代伊犁将军府设立水利厅，坎儿井数量达到历史高峰，吐鲁番地区超过千条。",
		},
		{
			Era:              "近现代",
			TimePeriod:       "1912年 - 2000年",
			KeyFeatures:      []string{"机械化施工", "混凝土衬砌", "水泵辅助", "现代测量技术"},
			Materials:        []string{"钢筋混凝土", "PVC管道", "土工膜", "钢材"},
			ConstructionTools: []string{"挖掘机", "盾构机", "全站仪", "钻探机"},
			AverageDepth:     80,
			AverageLength:    15.0,
			MaxFlowRate:      12000,
			WaterLossRate:    5,
			LabourRequirement: 3000,
			MaintenanceCycle: "传感器监测，智能维护",
			HistoricalNotes:  "新中国成立后，引入现代工程技术，但机井大量使用导致地下水位下降，坎儿井逐渐减少。",
		},
		{
			Era:              "当代数字化",
			TimePeriod:       "2000年至今",
			KeyFeatures:      []string{"智能监测", "BIM建模", "生态修复", "文化遗产保护"},
			Materials:        []string{"高性能混凝土", "纳米防渗材料", "物联网传感器", "光纤监测"},
			ConstructionTools: []string{"3D打印构件", "机器人巡检", "卫星遥感", "AI模拟系统"},
			AverageDepth:     100,
			AverageLength:    20.0,
			MaxFlowRate:      15000,
			WaterLossRate:    2,
			LabourRequirement: 500,
			MaintenanceCycle: "实时监测，预测性维护",
			HistoricalNotes:  "数字孪生技术与传统智慧结合，坎儿井作为文化遗产和生态工程获得新生。",
		},
	}

	innovations := []models.Innovation{
		{
			Name:        "竖井定位法",
			Era:         "魏晋南北朝",
			Description: "通过地面直线排列竖井确定暗渠走向，开创了地下工程的精确定位技术。",
			Impact:      8.5,
		},
		{
			Name:        "坡度控制技术",
			Era:         "隋唐时期",
			Description: "发明了利用水面水平原理控制暗渠纵坡的方法，确保水流畅通。",
			Impact:      9.0,
		},
		{
			Name:        "砖石衬砌工艺",
			Era:         "隋唐时期",
			Description: "用青砖和条石加固暗渠侧壁和拱顶，大幅延长工程寿命。",
			Impact:      8.8,
		},
		{
			Name:        "地下水库串联",
			Era:         "宋元时期",
			Description: "将多条坎儿井暗渠与地下蓄水层连通，形成调蓄系统。",
			Impact:      8.2,
		},
		{
			Name:        "冰窖调水法",
			Era:         "宋元时期",
			Description: "冬季储存冰雪融水于地下冰窖，夏季融化补充灌溉。",
			Impact:      7.5,
		},
		{
			Name:        "糯米灰浆防渗",
			Era:         "明清时期",
			Description: "用糯米汤混合石灰制成灰浆，防渗性能远超普通砂浆。",
			Impact:      9.2,
		},
		{
			Name:        "机械化开挖",
			Era:         "近现代",
			Description: "引入现代工程机械，施工效率提升数十倍。",
			Impact:      9.5,
		},
		{
			Name:        "物联网监测",
			Era:         "当代数字化",
			Description: "安装传感器实时监测水位、流量、沉降，实现智能管理。",
			Impact:      9.8,
		},
	}

	summary := "坎儿井技术历经两千多年演变，从简单的人工挖掘发展到数字化智能系统，反映了中华民族适应干旱环境的卓越智慧。每个时代的技术创新都建立在前人经验基础之上，体现了工程技术传承与创新的辩证关系。"

	return &models.TechnologyEvolutionAnalysis{
		Evolutions:     evolutions,
		KeyInnovations: innovations,
		Summary:        summary,
	}
}

func (s *CultureService) GetCrossEraComparison() *models.CrossEraComparison {
	karezMetrics := models.KarezComparisonMetrics{
		Name:              "传统坎儿井系统",
		WaterUseEfficiency: 85,
		EnergyConsumption: 0.5,
		SetupCostPerHa:    150000,
		MaintenanceCost:   3000,
		CropYieldBoost:    30,
		LifespanYears:     150,
		TechnologyLevel:   "古代智慧工程",
		EcosystemImpact:   "地下水自然引流，生态友好，维持绿洲生态平衡",
	}

	dripMetrics := models.DripIrrigationSystem{
		Name:              "现代滴灌系统",
		Description:       "基于压力管道和滴头的精准灌溉技术，配合水肥一体化管理。",
		WaterUseEfficiency: 95,
		EnergyConsumption: 8.0,
		SetupCostPerHa:    80000,
		MaintenanceCost:   12000,
		CropYieldBoost:    60,
		LifespanYears:     15,
		TechnologyLevel:   "现代农业高科技",
	}

	comparisons := []models.ComparisonItem{
		{
			Metric:         "水资源利用效率",
			KarezValue:     85,
			DripValue:      95,
			KarezUnit:      "%",
			DripUnit:       "%",
			BetterSolution: "滴灌",
			Notes:          "滴灌通过精准供水减少蒸发和渗漏，效率略高。但坎儿井无需额外水源加压，综合水损耗更低。",
		},
		{
			Metric:         "能源消耗",
			KarezValue:     0.5,
			DripValue:      8.0,
			KarezUnit:      "kWh/ha/天",
			DripUnit:       "kWh/ha/天",
			BetterSolution: "坎儿井",
			Notes:          "坎儿井利用重力自流，几乎不消耗能源；滴灌需要水泵加压和过滤系统，能耗显著。",
		},
		{
			Metric:         "初始建设成本",
			KarezValue:     150000,
			DripValue:      80000,
			KarezUnit:      "元/公顷",
			DripUnit:       "元/公顷",
			BetterSolution: "滴灌",
			Notes:          "坎儿井地下工程量大，初期投入高；滴灌系统安装便捷，短期成本更低。",
		},
		{
			Metric:         "年维护成本",
			KarezValue:     3000,
			DripValue:      12000,
			KarezUnit:      "元/公顷/年",
			DripUnit:       "元/公顷/年",
			BetterSolution: "坎儿井",
			Notes:          "坎儿井维护以清淤为主，成本低廉；滴灌易堵塞，需频繁更换滴头和维护设备。",
		},
		{
			Metric:         "作物增产幅度",
			KarezValue:     30,
			DripValue:      60,
			KarezUnit:      "%",
			DripUnit:       "%",
			BetterSolution: "滴灌",
			Notes:          "滴灌可配合水肥一体化精准控制，增产效果更显著；坎儿井为传统灌溉方式。",
		},
		{
			Metric:         "使用寿命",
			KarezValue:     150,
			DripValue:      15,
			KarezUnit:      "年",
			DripUnit:       "年",
			BetterSolution: "坎儿井",
			Notes:          "维护良好的坎儿井可使用百年以上；滴灌设备老化快，约10-15年需全面更新。",
		},
		{
			Metric:         "生态环境影响",
			KarezValue:     9.5,
			DripValue:      6.5,
			KarezUnit:      "评分",
			DripUnit:       "评分",
			BetterSolution: "坎儿井",
			Notes:          "坎儿井顺应自然水文循环，维持绿洲生态；滴灌抽取地下水加速水位下降，塑料废弃物污染。",
		},
		{
			Metric:         "气候适应性",
			KarezValue:     9.0,
			DripValue:      7.0,
			KarezUnit:      "评分",
			DripUnit:       "评分",
			BetterSolution: "坎儿井",
			Notes:          "坎儿井地下输水避免蒸发，极端高温下仍稳定运行；滴灌外露管道在极端气候下易损坏。",
		},
		{
			Metric:         "文化遗产价值",
			KarezValue:     10,
			DripValue:      1,
			KarezUnit:      "评分",
			DripUnit:       "评分",
			BetterSolution: "坎儿井",
			Notes:          "坎儿井是世界文化遗产，承载两千年人类智慧；滴灌为工业产品，无文化价值。",
		},
		{
			Metric:         "综合经济回报（50年周期）",
			KarezValue:     8.2,
			DripValue:      7.5,
			KarezUnit:      "评分",
			DripUnit:       "评分",
			BetterSolution: "坎儿井",
			Notes:          "长期来看，坎儿井低维护成本和长寿命带来更高经济回报，并附带文化旅游收益。",
		},
	}

	conclusion := "古代坎儿井与现代滴灌技术各有优势：滴灌在短期效率和增产方面表现突出，而坎儿井在可持续性、生态保护、文化价值和长期经济性方面具有不可替代的优势。最佳方案是两者结合——利用坎儿井的自然引水和生态功能，配合滴灌的精准灌溉技术，实现传统智慧与现代科技的完美融合。"

	return &models.CrossEraComparison{
		KarezSystem:       karezMetrics,
		DripIrrigation:    dripMetrics,
		ComparisonMetrics: comparisons,
		Conclusion:        conclusion,
	}
}
