package handlers

import (
	"karez-system/alarm_mqtt"
	"karez-system/db"
	dtureceiver "karez-system/dtu_receiver"
	hydraulicsim "karez-system/hydraulic_sim"
	karezculture "karez-system/karez_culture"
	"karez-system/models"
	waterlevel "karez-system/water_level"
	virtualdig "karez-system/virtual_dig"
	wateralloc "karez-system/water_allocator"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	database        *db.Database
	dtuReceiver     *dtureceiver.DtuReceiver
	hydraulicSim    *hydraulicsim.HydraulicSimulator
	waterAllocator  *wateralloc.WaterAllocator
	alarmManager    *alarmmqtt.AlarmManager
	cultureService  *karezculture.CultureService
	waterLevelSim   *waterlevel.WaterLevelSimulator
	virtualDigService *virtualdig.VirtualDigService
}

func New(database *db.Database,
	dtuReceiver *dtureceiver.DtuReceiver,
	hydraulicSim *hydraulicsim.HydraulicSimulator,
	waterAllocator *wateralloc.WaterAllocator,
	alarmManager *alarmmqtt.AlarmManager) *Handler {
	return &Handler{
		database:        database,
		dtuReceiver:     dtuReceiver,
		hydraulicSim:    hydraulicSim,
		waterAllocator:  waterAllocator,
		alarmManager:    alarmManager,
		cultureService:  karezculture.New(),
		waterLevelSim:   waterlevel.New(),
		virtualDigService: virtualdig.New(),
	}
}

type SensorDataRequest struct {
	KarezID         int     `json:"karez_id" binding:"required"`
	SegmentID       int     `json:"segment_id"`
	ShaftID         int     `json:"shaft_id"`
	SensorType      string  `json:"sensor_type" binding:"required"`
	SensorID        string  `json:"sensor_id" binding:"required"`
	FlowRate        float64 `json:"flow_rate"`
	WaterLevel      float64 `json:"water_level"`
	ShaftWaterLevel float64 `json:"shaft_water_level"`
	Evaporation     float64 `json:"evaporation"`
	Temperature     float64 `json:"temperature"`
	Turbidity       float64 `json:"turbidity"`
	Velocity        float64 `json:"velocity"`
}

func (h *Handler) PostSensorData(c *gin.Context) {
	var req SensorDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data := &models.SensorData{
		Time:            time.Now(),
		KarezID:         req.KarezID,
		SegmentID:       req.SegmentID,
		ShaftID:         req.ShaftID,
		SensorType:      req.SensorType,
		SensorID:        req.SensorID,
		FlowRate:        req.FlowRate,
		WaterLevel:      req.WaterLevel,
		ShaftWaterLevel: req.ShaftWaterLevel,
		Evaporation:     req.Evaporation,
		Temperature:     req.Temperature,
		Turbidity:       req.Turbidity,
		Velocity:        req.Velocity,
	}

	if err := h.dtuReceiver.ReceiveHTTP(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Sensor data received"})
}

func (h *Handler) GetKarezSystems(c *gin.Context) {
	systems, err := h.database.GetKarezSystems(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, systems)
}

func (h *Handler) GetAqueductSegments(c *gin.Context) {
	karezID, err := strconv.Atoi(c.Param("karez_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid karez_id"})
		return
	}

	segments, err := h.database.GetAqueductSegments(c.Request.Context(), karezID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, segments)
}

func (h *Handler) GetVerticalShafts(c *gin.Context) {
	karezID, err := strconv.Atoi(c.Param("karez_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid karez_id"})
		return
	}

	shafts, err := h.database.GetVerticalShafts(c.Request.Context(), karezID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shafts)
}

func (h *Handler) GetBranchChannels(c *gin.Context) {
	karezID, err := strconv.Atoi(c.Param("karez_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid karez_id"})
		return
	}

	channels, err := h.database.GetBranchChannels(c.Request.Context(), karezID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, channels)
}

func (h *Handler) GetOases(c *gin.Context) {
	karezID, err := strconv.Atoi(c.Param("karez_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid karez_id"})
		return
	}

	oases, err := h.database.GetOases(c.Request.Context(), karezID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, oases)
}

func (h *Handler) GetLatestSensorData(c *gin.Context) {
	karezID, err := strconv.Atoi(c.Param("karez_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid karez_id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	data, err := h.database.GetLatestSensorData(c.Request.Context(), karezID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) GetSensorDataByRange(c *gin.Context) {
	karezID, err := strconv.Atoi(c.Param("karez_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid karez_id"})
		return
	}

	startTimeStr := c.Query("start")
	endTimeStr := c.Query("end")

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		startTime = time.Now().Add(-24 * time.Hour)
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		endTime = time.Now()
	}

	data, err := h.database.GetSensorDataByRange(c.Request.Context(), karezID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

type SimulateRequest struct {
	KarezID int `json:"karez_id" binding:"required"`
}

func (h *Handler) RunSimulation(c *gin.Context) {
	var req SimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.hydraulicSim.RunFullSimulation(c.Request.Context(), req.KarezID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Simulation completed"})
}

type AllocateRequest struct {
	KarezID            int     `json:"karez_id" binding:"required"`
	TotalAvailableFlow float64 `json:"total_available_flow" binding:"required"`
}

func (h *Handler) RunAllocation(c *gin.Context) {
	var req AllocateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	solution, err := h.waterAllocator.OptimizeAllocation(c.Request.Context(), req.KarezID, req.TotalAvailableFlow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"objective_value": solution.ObjectiveValue,
		"total_allocated": solution.TotalAllocated,
		"allocations":     solution.Allocations,
		"demand_met":      solution.DemandMet,
	})
}

func (h *Handler) GetActiveAlerts(c *gin.Context) {
	karezID, err := strconv.Atoi(c.Param("karez_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid karez_id"})
		return
	}

	alerts, err := h.database.GetActiveAlerts(c.Request.Context(), karezID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

type AcknowledgeAlertRequest struct {
	AlertID int `json:"alert_id" binding:"required"`
}

func (h *Handler) AcknowledgeAlert(c *gin.Context) {
	var req AcknowledgeAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.alarmManager.AcknowledgeAlert(c.Request.Context(), req.AlertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

type ResolveAlertRequest struct {
	AlertID int `json:"alert_id" binding:"required"`
}

func (h *Handler) ResolveAlert(c *gin.Context) {
	var req ResolveAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.alarmManager.ResolveAlert(c.Request.Context(), req.AlertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *Handler) CheckAlerts(c *gin.Context) {
	karezID, err := strconv.Atoi(c.Param("karez_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid karez_id"})
		return
	}

	if err := h.alarmManager.CheckAndAlert(c.Request.Context(), karezID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Alert check completed"})
}

type SimulateHydraulicRequest struct {
	Width                float64 `json:"width" binding:"required"`
	Height               float64 `json:"height" binding:"required"`
	Slope                float64 `json:"slope" binding:"required"`
	RoughnessCoeff       float64 `json:"roughness_coeff"`
	SeepageCoeff         float64 `json:"seepage_coeff"`
	SoilType             string  `json:"soil_type"`
	SoilCorrectionFactor float64 `json:"soil_correction_factor"`
	Length               float64 `json:"length" binding:"required"`
	InflowRate           float64 `json:"inflow_rate" binding:"required"`
	Temperature          float64 `json:"temperature"`
}

func (h *Handler) SimulateHydraulic(c *gin.Context) {
	var req SimulateHydraulicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RoughnessCoeff == 0 {
		req.RoughnessCoeff = 0.013
	}
	if req.SeepageCoeff == 0 {
		req.SeepageCoeff = 0.0001
	}
	if req.SoilType == "" {
		req.SoilType = "gravel"
	}
	if req.SoilCorrectionFactor == 0 {
		req.SoilCorrectionFactor = 1.0
	}
	if req.Temperature == 0 {
		req.Temperature = 25.0
	}

	params := hydraulicsim.ChannelParams{
		Width:                req.Width,
		Height:               req.Height,
		Slope:                req.Slope,
		RoughnessCoeff:       req.RoughnessCoeff,
		SeepageCoeff:         req.SeepageCoeff,
		SoilType:             req.SoilType,
		SoilCorrectionFactor: req.SoilCorrectionFactor,
		Length:               req.Length,
		Temperature:          req.Temperature,
	}

	result := h.hydraulicSim.SimulateSegmentDirect(params, req.InflowRate)
	sedimentationRisk := h.hydraulicSim.EstimateSedimentationRisk(result.FlowVelocity)

	c.JSON(http.StatusOK, gin.H{
		"inflow_rate":        result.InflowRate,
		"outflow_rate":       result.OutflowRate,
		"seepage_loss":       result.SeepageLoss,
		"evaporation_loss":   result.EvaporationLoss,
		"total_loss":         result.TotalLoss,
		"water_depth":        result.WaterDepth,
		"flow_velocity":      result.FlowVelocity,
		"reynolds_number":    result.ReynoldsNumber,
		"froude_number":      result.FroudeNumber,
		"head_loss":          result.HeadLoss,
		"sedimentation_risk": sedimentationRisk,
	})
}

func (h *Handler) GetDashboardData(c *gin.Context) {
	karezID, err := strconv.Atoi(c.Param("karez_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid karez_id"})
		return
	}

	ctx := c.Request.Context()

	segments, _ := h.database.GetAqueductSegments(ctx, karezID)
	shafts, _ := h.database.GetVerticalShafts(ctx, karezID)
	channels, _ := h.database.GetBranchChannels(ctx, karezID)
	oases, _ := h.database.GetOases(ctx, karezID)
	alerts, _ := h.database.GetActiveAlerts(ctx, karezID)
	sensorData, _ := h.database.GetLatestSensorData(ctx, karezID, 50)

	totalFlow := 0.0
	latestBySegment := make(map[int]models.SensorData)
	for _, d := range sensorData {
		if d.SensorType == "flow" {
			if existing, ok := latestBySegment[d.SegmentID]; !ok || d.Time.After(existing.Time) {
				latestBySegment[d.SegmentID] = d
			}
		}
	}

	for _, d := range latestBySegment {
		if d.FlowRate > totalFlow {
			totalFlow = d.FlowRate
		}
	}

	totalDemand := 0.0
	for _, o := range oases {
		totalDemand += o.DailyWaterDemand / 86400.0
	}

	supplyRatio := 0.0
	if totalDemand > 0 {
		supplyRatio = totalFlow / totalDemand
	}

	c.JSON(http.StatusOK, gin.H{
		"karez_id":       karezID,
		"segments":       segments,
		"shafts":         shafts,
		"branch_channels": channels,
		"oases":          oases,
		"active_alerts":  alerts,
		"latest_data":    sensorData,
		"total_flow":     totalFlow,
		"total_demand":   totalDemand,
		"supply_ratio":   supplyRatio,
		"alert_count":    len(alerts),
	})
}

func (h *Handler) GetTechnologyEvolution(c *gin.Context) {
	analysis := h.cultureService.GetTechnologyEvolution()
	c.JSON(http.StatusOK, analysis)
}

func (h *Handler) GetCrossEraComparison(c *gin.Context) {
	comparison := h.cultureService.GetCrossEraComparison()
	c.JSON(http.StatusOK, comparison)
}

func (h *Handler) GetWaterLevelScenarios(c *gin.Context) {
	scenarios := h.waterLevelSim.GetDefaultScenarios()
	c.JSON(http.StatusOK, scenarios)
}

func (h *Handler) SimulateWaterLevelImpact(c *gin.Context) {
	var req models.WaterLevelSimulationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results := h.waterLevelSim.SimulateWaterLevelImpact(req)
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"results":  results,
		"count":    len(results),
	})
}

func (h *Handler) GetDefaultTerrain(c *gin.Context) {
	terrain := h.virtualDigService.GetDefaultTerrain()
	c.JSON(http.StatusOK, terrain)
}

func (h *Handler) SaveVirtualDigProject(c *gin.Context) {
	var req models.VirtualDigSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.virtualDigService.SaveProject(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"project": project,
	})
}

func (h *Handler) SimulateVirtualDig(c *gin.Context) {
	var req models.VirtualDigSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.virtualDigService.SimulateDesign(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"project": project,
	})
}

func (h *Handler) GetVirtualDigProject(c *gin.Context) {
	id := c.Param("id")
	project, exists := h.virtualDigService.GetProject(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, project)
}

func (h *Handler) ListVirtualDigProjects(c *gin.Context) {
	projects := h.virtualDigService.ListProjects()
	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"count":    len(projects),
	})
}

func (h *Handler) DeleteVirtualDigProject(c *gin.Context) {
	id := c.Param("id")
	if ok := h.virtualDigService.DeleteProject(id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Project deleted"})
}

func (h *Handler) GetDigGuide(c *gin.Context) {
	guide := h.virtualDigService.GetDigGuide()
	c.JSON(http.StatusOK, guide)
}

func (h *Handler) GetDesignTemplates(c *gin.Context) {
	templates := h.virtualDigService.GetDesignTemplates()
	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"count":     len(templates),
	})
}

func (h *Handler) GetQuickTips(c *gin.Context) {
	tips := h.virtualDigService.GetQuickTips()
	c.JSON(http.StatusOK, gin.H{
		"tips":  tips,
		"count": len(tips),
	})
}
