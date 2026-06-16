package models

import "time"

type KarezSystem struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Location       string    `json:"location"`
	TotalLength    float64   `json:"total_length"`
	HeadElevation  float64   `json:"head_elevation"`
	TailElevation  float64   `json:"tail_elevation"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
}

type AqueductSegment struct {
	ID                   int       `json:"id"`
	KarezID              int       `json:"karez_id"`
	SegmentName          string    `json:"segment_name"`
	SegmentOrder         int       `json:"segment_order"`
	StartElevation       float64   `json:"start_elevation"`
	EndElevation         float64   `json:"end_elevation"`
	Length               float64   `json:"length"`
	Width                float64   `json:"width"`
	Height               float64   `json:"height"`
	Slope                float64   `json:"slope"`
	RoughnessCoeff       float64   `json:"roughness_coeff"`
	SeepageCoeff         float64   `json:"seepage_coeff"`
	SoilType             string    `json:"soil_type"`
	SoilCorrectionFactor float64   `json:"soil_correction_factor"`
	IsMainChannel        bool      `json:"is_main_channel"`
	CreatedAt            time.Time `json:"created_at"`
}

type VerticalShaft struct {
	ID               int       `json:"id"`
	KarezID          int       `json:"karez_id"`
	SegmentID        int       `json:"segment_id"`
	ShaftName        string    `json:"shaft_name"`
	ShaftOrder       int       `json:"shaft_order"`
	GroundElevation  float64   `json:"ground_elevation"`
	ShaftDepth       float64   `json:"shaft_depth"`
	Diameter         float64   `json:"diameter"`
	DistanceFromHead float64   `json:"distance_from_head"`
	CreatedAt        time.Time `json:"created_at"`
}

type BranchChannel struct {
	ID                int       `json:"id"`
	KarezID           int       `json:"karez_id"`
	MainSegmentID     int       `json:"main_segment_id"`
	BranchName        string    `json:"branch_name"`
	DesignFlow        float64   `json:"design_flow"`
	MaxFlow           float64   `json:"max_flow"`
	Length            float64   `json:"length"`
	Width             float64   `json:"width"`
	Height            float64   `json:"height"`
	CurrentAllocation float64   `json:"current_allocation"`
	CreatedAt         time.Time `json:"created_at"`
}

type Oasis struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	BranchChannelID  int       `json:"branch_channel_id"`
	Area             float64   `json:"area"`
	DailyWaterDemand float64   `json:"daily_water_demand"`
	Priority         int       `json:"priority"`
	CreatedAt        time.Time `json:"created_at"`
}

type SensorData struct {
	Time            time.Time `json:"time"`
	KarezID         int       `json:"karez_id"`
	SegmentID       int       `json:"segment_id,omitempty"`
	ShaftID         int       `json:"shaft_id,omitempty"`
	SensorType      string    `json:"sensor_type"`
	SensorID        string    `json:"sensor_id"`
	FlowRate        float64   `json:"flow_rate,omitempty"`
	WaterLevel      float64   `json:"water_level,omitempty"`
	ShaftWaterLevel float64   `json:"shaft_water_level,omitempty"`
	Evaporation     float64   `json:"evaporation,omitempty"`
	Temperature     float64   `json:"temperature,omitempty"`
	Turbidity       float64   `json:"turbidity,omitempty"`
	Velocity        float64   `json:"velocity,omitempty"`
	Metadata        string    `json:"metadata,omitempty"`
}

type SimulationResult struct {
	Time            time.Time `json:"time"`
	KarezID         int       `json:"karez_id"`
	SegmentID       int       `json:"segment_id,omitempty"`
	SimulationType  string    `json:"simulation_type"`
	InflowRate      float64   `json:"inflow_rate"`
	OutflowRate     float64   `json:"outflow_rate"`
	SeepageLoss     float64   `json:"seepage_loss"`
	EvaporationLoss float64   `json:"evaporation_loss"`
	TotalLoss       float64   `json:"total_loss"`
	WaterDepth      float64   `json:"water_depth"`
	FlowVelocity    float64   `json:"flow_velocity"`
	ReynoldsNumber  float64   `json:"reynolds_number"`
	FroudeNumber    float64   `json:"froude_number"`
	HeadLoss        float64   `json:"head_loss"`
	Metadata        string    `json:"metadata,omitempty"`
}

type AllocationResult struct {
	Time               time.Time `json:"time"`
	KarezID            int       `json:"karez_id"`
	BranchChannelID    int       `json:"branch_channel_id,omitempty"`
	OasisID            int       `json:"oasis_id,omitempty"`
	AllocatedFlow      float64   `json:"allocated_flow"`
	AllocationRatio    float64   `json:"allocation_ratio"`
	DemandMet          float64   `json:"demand_met"`
	OptimizationMethod string    `json:"optimization_method"`
	ObjectiveValue     float64   `json:"objective_value"`
	Metadata           string    `json:"metadata,omitempty"`
}

type AlertEvent struct {
	Time            time.Time `json:"time"`
	AlertID         int       `json:"alert_id"`
	KarezID         int       `json:"karez_id,omitempty"`
	SegmentID       int       `json:"segment_id,omitempty"`
	BranchChannelID int       `json:"branch_channel_id,omitempty"`
	AlertType       string    `json:"alert_type"`
	AlertLevel      string    `json:"alert_level"`
	Message         string    `json:"message"`
	CurrentValue    float64   `json:"current_value"`
	ThresholdValue  float64   `json:"threshold_value"`
	Acknowledged    bool      `json:"acknowledged"`
	Resolved        bool      `json:"resolved"`
	Metadata        string    `json:"metadata,omitempty"`
}

type AlertRule struct {
	ID         int       `json:"id"`
	RuleName   string    `json:"rule_name"`
	RuleType   string    `json:"rule_type"`
	Threshold  float64   `json:"threshold"`
	Comparison string    `json:"comparison"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
}

type EraTechnology struct {
	Era              string   `json:"era"`
	TimePeriod       string   `json:"time_period"`
	KeyFeatures      []string `json:"key_features"`
	Materials        []string `json:"materials"`
	ConstructionTools []string `json:"construction_tools"`
	AverageDepth     float64  `json:"average_depth_meters"`
	AverageLength    float64  `json:"average_length_km"`
	MaxFlowRate      float64  `json:"max_flow_rate_m3_per_day"`
	WaterLossRate    float64  `json:"water_loss_rate_percent"`
	LabourRequirement float64 `json:"labour_requirement_person_days_per_km"`
	MaintenanceCycle string   `json:"maintenance_cycle"`
	HistoricalNotes  string   `json:"historical_notes"`
}

type TechnologyEvolutionAnalysis struct {
	Evolutions     []EraTechnology `json:"evolutions"`
	KeyInnovations []Innovation    `json:"key_innovations"`
	Summary        string          `json:"summary"`
}

type Innovation struct {
	Name        string  `json:"name"`
	Era         string  `json:"era"`
	Description string  `json:"description"`
	Impact      float64 `json:"impact_score"`
}

type DripIrrigationSystem struct {
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	WaterUseEfficiency float64 `json:"water_use_efficiency_percent"`
	EnergyConsumption float64 `json:"energy_consumption_kwh_per_ha_per_day"`
	SetupCostPerHa    float64 `json:"setup_cost_per_ha"`
	MaintenanceCost   float64 `json:"maintenance_cost_per_ha_per_year"`
	CropYieldBoost    float64 `json:"crop_yield_boost_percent"`
	LifespanYears     int     `json:"lifespan_years"`
	TechnologyLevel   string  `json:"technology_level"`
}

type CrossEraComparison struct {
	KarezSystem       KarezComparisonMetrics `json:"karez_system"`
	DripIrrigation    DripIrrigationSystem   `json:"drip_irrigation"`
	ComparisonMetrics []ComparisonItem       `json:"comparison_metrics"`
	Conclusion        string                 `json:"conclusion"`
}

type KarezComparisonMetrics struct {
	Name              string  `json:"name"`
	WaterUseEfficiency float64 `json:"water_use_efficiency_percent"`
	EnergyConsumption float64 `json:"energy_consumption_kwh_per_ha_per_day"`
	SetupCostPerHa    float64 `json:"setup_cost_per_ha"`
	MaintenanceCost   float64 `json:"maintenance_cost_per_ha_per_year"`
	CropYieldBoost    float64 `json:"crop_yield_boost_percent"`
	LifespanYears     int     `json:"lifespan_years"`
	TechnologyLevel   string  `json:"technology_level"`
	EcosystemImpact   string  `json:"ecosystem_impact"`
}

type ComparisonItem struct {
	Metric         string  `json:"metric"`
	KarezValue     float64 `json:"karez_value"`
	DripValue      float64 `json:"drip_value"`
	KarezUnit      string  `json:"karez_unit"`
	DripUnit       string  `json:"drip_unit"`
	BetterSolution string  `json:"better_solution"`
	Notes          string  `json:"notes"`
}

type WaterLevelScenario struct {
	ScenarioName     string    `json:"scenario_name"`
	InitialWaterLevel float64   `json:"initial_water_level_meters"`
	TargetWaterLevel  float64   `json:"target_water_level_meters"`
	ChangeRate        float64   `json:"change_rate_m_per_year"`
	DurationYears     int       `json:"duration_years"`
	Description       string    `json:"description"`
}

type WaterLevelSimulationRequest struct {
	KarezID          int                 `json:"karez_id" binding:"required"`
	Scenarios        []WaterLevelScenario `json:"scenarios"`
	BaselineFlowRate float64             `json:"baseline_flow_rate_m3_per_day"`
	ShaftDepth       float64             `json:"shaft_depth_meters"`
	AquiferType      string              `json:"aquifer_type"`
}

type WaterLevelDataPoint struct {
	Year                  int     `json:"year"`
	WaterLevel            float64 `json:"water_level_meters"`
	FlowRate              float64 `json:"flow_rate_m3_per_day"`
	FlowChangePercent     float64 `json:"flow_change_percent"`
	ShaftIntakeDepth      float64 `json:"shaft_intake_depth_meters"`
	IsFlowSustained       bool    `json:"is_flow_sustained"`
	WarningLevel          string  `json:"warning_level"`
}

type WaterLevelSimulationResult struct {
	KarezID        int                   `json:"karez_id"`
	ScenarioName   string                `json:"scenario_name"`
	BaselineFlow   float64               `json:"baseline_flow_rate_m3_per_day"`
	DataPoints     []WaterLevelDataPoint `json:"data_points"`
	FinalFlowRate  float64               `json:"final_flow_rate_m3_per_day"`
	TotalDecline   float64               `json:"total_decline_percent"`
	YearsUntilDry  int                   `json:"years_until_dry"`
	Recommendations []string             `json:"recommendations"`
}

type VirtualDigProject struct {
	ID            string              `json:"id"`
	ProjectName   string              `json:"project_name"`
	Creator       string              `json:"creator"`
	CreatedAt     time.Time           `json:"created_at"`
	TerrainMap    TerrainConfig       `json:"terrain_map"`
	Channels      []DigChannel        `json:"channels"`
	Shafts        []DigShaft          `json:"shafts"`
	Statistics    DigStatistics       `json:"statistics"`
	SimulatedFlow float64             `json:"simulated_flow_m3_per_day"`
	Feasibility   FeasibilityReport   `json:"feasibility"`
}

type TerrainConfig struct {
	WidthKm         float64       `json:"width_km"`
	LengthKm        float64       `json:"length_km"`
	HeadElevation   float64       `json:"head_elevation_meters"`
	TailElevation   float64       `json:"tail_elevation_meters"`
	WaterTableDepth float64       `json:"water_table_depth_meters"`
	SoilType        string        `json:"soil_type"`
	Obstacles       []MapObstacle `json:"obstacles"`
}

type MapObstacle struct {
	ID       string  `json:"id"`
	Type     string  `json:"type"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Radius   float64 `json:"radius"`
	Label    string  `json:"label"`
}

type DigChannel struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	IsMain      bool        `json:"is_main"`
	Points      []GeoPoint  `json:"points"`
	Width       float64     `json:"width_meters"`
	Height      float64     `json:"height_meters"`
	Depth       float64     `json:"depth_meters"`
	Slope       float64     `json:"slope"`
	Length      float64     `json:"length_meters"`
	ParentID    string      `json:"parent_id,omitempty"`
}

type GeoPoint struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Elevation float64 `json:"elevation"`
}

type DigShaft struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	ChannelID        string  `json:"channel_id"`
	Position         GeoPoint `json:"position"`
	Depth            float64 `json:"depth_meters"`
	Diameter         float64 `json:"diameter_meters"`
	ReachesWater     bool    `json:"reaches_water_table"`
	DistanceFromHead float64 `json:"distance_from_head_meters"`
}

type DigStatistics struct {
	TotalChannelLength   float64 `json:"total_channel_length_meters"`
	TotalShafts          int     `json:"total_shafts"`
	TotalExcavationVolume float64 `json:"total_excavation_volume_m3"`
	EstimatedManDays     float64 `json:"estimated_man_days"`
	EstimatedCost        float64 `json:"estimated_cost"`
	AverageDepth         float64 `json:"average_depth_meters"`
}

type FeasibilityReport struct {
	IsFeasible     bool               `json:"is_feasible"`
	OverallScore   float64            `json:"overall_score"`
	HydraulicScore float64            `json:"hydraulic_score"`
	GeologicalScore float64           `json:"geological_score"`
	EconomicScore  float64            `json:"economic_score"`
	Issues         []FeasibilityIssue `json:"issues"`
	Suggestions    []string           `json:"suggestions"`
}

type FeasibilityIssue struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Location string `json:"location,omitempty"`
}

type VirtualDigSaveRequest struct {
	ProjectName string        `json:"project_name" binding:"required"`
	Creator     string        `json:"creator"`
	TerrainMap  TerrainConfig `json:"terrain_map" binding:"required"`
	Channels    []DigChannel  `json:"channels"`
	Shafts      []DigShaft    `json:"shafts"`
}

type VirtualDigSimulateRequest struct {
	TerrainMap  TerrainConfig `json:"terrain_map" binding:"required"`
	Channels    []DigChannel  `json:"channels"`
	Shafts      []DigShaft    `json:"shafts"`
}
