package main

import (
	"context"
	"karez-system/alarm_mqtt"
	"karez-system/config"
	"karez-system/db"
	"karez-system/dtu_receiver"
	"karez-system/handlers"
	"karez-system/hydraulic_sim"
	"karez-system/metrics"
	"karez-system/models"
	"karez-system/mqtt"
	"karez-system/water_allocator"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	database, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("Database connected successfully")

	mqttClient, err := mqtt.New(cfg)
	if err != nil {
		log.Printf("Warning: Failed to connect to MQTT broker: %v", err)
		log.Println("Continuing without MQTT support")
		mqttClient = nil
	} else {
		defer mqttClient.Close()
		log.Println("MQTT connected successfully")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sensorDataChan := make(chan *models.SensorData, 100)
	simRequestChan := make(chan hydraulicsim.SimRequest, 10)
	simResultChan := make(chan hydraulicsim.SimResult, 20)
	allocRequestChan := make(chan wateralloc.AllocationRequest, 10)
	allocResultChan := make(chan wateralloc.AllocationResponse, 10)
	alarmRequestChan := make(chan alarmmqtt.AlarmCheckRequest, 10)

	dtuReceiver := dtureceiver.New(cfg, database, mqttClient, sensorDataChan)
	if err := dtuReceiver.Start(ctx); err != nil {
		log.Printf("Warning: DTU receiver start failed: %v", err)
	}

	hydraulicSim := hydraulicsim.New(cfg, database, simRequestChan, simResultChan)
	hydraulicSim.Start(ctx)

	waterAllocator := wateralloc.New(cfg, database, allocRequestChan, allocResultChan)
	waterAllocator.Start(ctx)

	alarmManager := alarmmqtt.New(cfg, database, mqttClient, alarmRequestChan)
	alarmManager.Start(ctx)

	h := handlers.New(database, dtuReceiver, hydraulicSim, waterAllocator, alarmManager)

	go func() {
		log.Printf("Pprof server starting on port 6060")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Printf("Pprof server failed: %v", err)
		}
	}()

	r := gin.Default()

	r.Use(metrics.PrometheusMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	r.GET("/metrics", gin.WrapH(metrics.PrometheusHandler()))

	api := r.Group("/api")
	{
		api.GET("/karez", h.GetKarezSystems)
		api.GET("/karez/:karez_id/segments", h.GetAqueductSegments)
		api.GET("/karez/:karez_id/shafts", h.GetVerticalShafts)
		api.GET("/karez/:karez_id/branches", h.GetBranchChannels)
		api.GET("/karez/:karez_id/oases", h.GetOases)
		api.GET("/karez/:karez_id/dashboard", h.GetDashboardData)

		api.POST("/sensor", h.PostSensorData)
		api.GET("/sensor/:karez_id/latest", h.GetLatestSensorData)
		api.GET("/sensor/:karez_id/range", h.GetSensorDataByRange)

		api.POST("/simulate", h.RunSimulation)
		api.POST("/simulate/hydraulic", h.SimulateHydraulic)

		api.POST("/allocate", h.RunAllocation)

		api.GET("/alerts/:karez_id", h.GetActiveAlerts)
		api.POST("/alerts/check/:karez_id", h.CheckAlerts)
		api.POST("/alerts/acknowledge", h.AcknowledgeAlert)
		api.POST("/alerts/resolve", h.ResolveAlert)

		api.GET("/culture/evolution", h.GetTechnologyEvolution)
		api.GET("/culture/comparison", h.GetCrossEraComparison)

		api.GET("/water-level/scenarios", h.GetWaterLevelScenarios)
		api.POST("/water-level/simulate", h.SimulateWaterLevelImpact)

		api.GET("/virtual-dig/terrain", h.GetDefaultTerrain)
		api.GET("/virtual-dig/guide", h.GetDigGuide)
		api.GET("/virtual-dig/templates", h.GetDesignTemplates)
		api.GET("/virtual-dig/tips", h.GetQuickTips)
		api.GET("/virtual-dig/projects", h.ListVirtualDigProjects)
		api.GET("/virtual-dig/projects/:id", h.GetVirtualDigProject)
		api.POST("/virtual-dig/projects", h.SaveVirtualDigProject)
		api.POST("/virtual-dig/simulate", h.SimulateVirtualDig)
		api.DELETE("/virtual-dig/projects/:id", h.DeleteVirtualDigProject)
	}

	go startPeriodicTasks(cfg, simRequestChan, alarmRequestChan)
	go startMetricsCollector(sensorDataChan, simRequestChan, simResultChan,
		allocRequestChan, allocResultChan, alarmRequestChan)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := r.Run(":" + cfg.ServerPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)
	cancel()
	time.Sleep(500 * time.Millisecond)
	log.Println("Server stopped")
}

func startPeriodicTasks(cfg *config.Config,
	simRequestChan chan<- hydraulicsim.SimRequest,
	alarmRequestChan chan<- alarmmqtt.AlarmCheckRequest) {

	simTicker := time.NewTicker(time.Duration(cfg.SimulationInterval) * time.Second)
	alertTicker := time.NewTicker(time.Duration(cfg.AlertCheckInterval) * time.Second)

	defer simTicker.Stop()
	defer alertTicker.Stop()

	for {
		select {
		case <-simTicker.C:
			select {
			case simRequestChan <- hydraulicsim.SimRequest{KarezID: 1}:
			default:
				log.Println("Periodic sim: channel full, skipping")
			}

		case <-alertTicker.C:
			select {
			case alarmRequestChan <- alarmmqtt.AlarmCheckRequest{KarezID: 1}:
			default:
				log.Println("Periodic alarm: channel full, skipping")
			}
		}
	}
}

func startMetricsCollector(
	sensorDataChan chan *models.SensorData,
	simRequestChan chan hydraulicsim.SimRequest,
	simResultChan chan hydraulicsim.SimResult,
	allocRequestChan chan wateralloc.AllocationRequest,
	allocResultChan chan wateralloc.AllocationResponse,
	alarmRequestChan chan alarmmqtt.AlarmCheckRequest,
) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics.SetChannelBackpressure("sensor_data", float64(len(sensorDataChan))/float64(cap(sensorDataChan)))
		metrics.SetChannelBackpressure("sim_request", float64(len(simRequestChan))/float64(cap(simRequestChan)))
		metrics.SetChannelBackpressure("sim_result", float64(len(simResultChan))/float64(cap(simResultChan)))
		metrics.SetChannelBackpressure("alloc_request", float64(len(allocRequestChan))/float64(cap(allocRequestChan)))
		metrics.SetChannelBackpressure("alloc_result", float64(len(allocResultChan))/float64(cap(allocResultChan)))
		metrics.SetChannelBackpressure("alarm_request", float64(len(alarmRequestChan))/float64(cap(alarmRequestChan)))
	}
}
