package sensor

import (
	"sync"
	"time"
)

// SensorType represents different types of sensors
type SensorType string

const (
	// different sensor types blyat
	TypeTouch    SensorType = "touch"
	TypePressure SensorType = "pressure"
	TypeMotion   SensorType = "motion"
	TypeTemp     SensorType = "temperature"
)

// SensorData represents data from single sensor
type SensorData struct {
	Type      SensorType
	Value     float64
	Timestamp time.Time
}

// Hub manages all sensor systems
type Hub struct {
	sensors map[SensorType][]float64
	mu      sync.RWMutex
	
	// channels for sensor data
	dataChan chan SensorData
	done     chan struct{}
}

// NewHub creates new sensor management system
func NewHub() (*Hub, error) {
	hub := &Hub{
		sensors:  make(map[SensorType][]float64),
		dataChan: make(chan SensorData, 100),
		done:     make(chan struct{}),
	}
	
	// initialize sensor types
	hub.sensors[TypeTouch] = make([]float64, 0)
	hub.sensors[TypePressure] = make([]float64, 0)
	hub.sensors[TypeMotion] = make([]float64, 0)
	hub.sensors[TypeTemp] = make([]float64, 0)
	
	go hub.processData()
	
	return hub, nil
}

// processData handles incoming sensor data
func (h *Hub) processData() {
	for {
		select {
		case data := <-h.dataChan:
			h.mu.Lock()
			h.sensors[data.Type] = append(h.sensors[data.Type], data.Value)
			// keep only last 1000 readings
			if len(h.sensors[data.Type]) > 1000 {
				h.sensors[data.Type] = h.sensors[data.Type][1:]
			}
			h.mu.Unlock()
		case <-h.done:
			return
		}
	}
}

// AddSensorData adds new sensor reading
func (h *Hub) AddSensorData(data SensorData) {
	h.dataChan <- data
}

// GetSensorData returns latest sensor readings
func (h *Hub) GetSensorData(sType SensorType) []float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if data, ok := h.sensors[sType]; ok {
		return data
	}
	return nil
}

// Shutdown stops sensor processing
func (h *Hub) Shutdown() {
	close(h.done)
	close(h.dataChan)
} 