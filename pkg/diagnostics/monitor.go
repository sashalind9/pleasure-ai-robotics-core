package diagnostics

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/sashalind/sex-artifical-intelligence/pkg/core"
)

// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	Timestamp     time.Time `json:"timestamp"`
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryUsage   float64   `json:"memory_usage"`
	Temperature   float64   `json:"temperature"`
	UptimeSeconds int64     `json:"uptime_seconds"`
}

// Monitor handles system diagnostics
type Monitor struct {
	system   *core.System
	mu       sync.RWMutex
	
	// diagnostic data
	metrics  []SystemMetrics
	logFile  *os.File
}

// StartMonitoring initializes diagnostic monitoring
func StartMonitoring(sys *core.System) error {
	logFile, err := os.OpenFile("diagnostics.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	
	monitor := &Monitor{
		system:  sys,
		metrics: make([]SystemMetrics, 0),
		logFile: logFile,
	}
	
	go monitor.collectMetrics()
	return nil
}

// collectMetrics gathers system performance data
func (m *Monitor) collectMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		if !m.system.IsActive() {
			m.logFile.Close()
			return
		}
		
		metrics := m.gatherMetrics()
		m.saveMetrics(metrics)
	}
}

// gatherMetrics collects current system metrics
func (m *Monitor) gatherMetrics() SystemMetrics {
	// TODO: implement actual metric collection
	// For now return dummy data
	return SystemMetrics{
		Timestamp:     time.Now(),
		CPUUsage:      45.5,
		MemoryUsage:   1024.5,
		Temperature:   37.2,
		UptimeSeconds: int64(m.system.GetUptime().Seconds()),
	}
}

// saveMetrics saves metrics to log file
func (m *Monitor) saveMetrics(metrics SystemMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.metrics = append(m.metrics, metrics)
	
	// keep only last 1000 metrics
	if len(m.metrics) > 1000 {
		m.metrics = m.metrics[1:]
	}
	
	// save to log file
	data, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Failed to marshal metrics: %v", err)
		return
	}
	
	if _, err := m.logFile.WriteString(string(data) + "\n"); err != nil {
		log.Printf("Failed to write metrics: %v", err)
	}
}

// GetLatestMetrics returns most recent system metrics
func (m *Monitor) GetLatestMetrics() *SystemMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.metrics) == 0 {
		return nil
	}
	
	latest := m.metrics[len(m.metrics)-1]
	return &latest
} 