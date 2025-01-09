package core

import (
	"context"
	"sync"
	"time"

	"github.com/sashalind/sex-artifical-intelligence/pkg/neural"
	"github.com/sashalind/sex-artifical-intelligence/pkg/sensor"
)

// System represents main control system blyat
type System struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	
	neuralNet  *neural.Network
	sensorHub  *sensor.Hub
	
	// mutex for thread safety, like in soviet russia
	mu         sync.RWMutex
	
	// system states
	isActive   bool
	startTime  time.Time
}

// NewSystem creates new instance of our glorious system
func NewSystem() (*System, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	neuralNet, err := neural.NewNetwork()
	if err != nil {
		cancel()
		return nil, err
	}
	
	sensorHub, err := sensor.NewHub()
	if err != nil {
		cancel()
		return nil, err
	}
	
	return &System{
		ctx:        ctx,
		cancelFunc: cancel,
		neuralNet:  neuralNet,
		sensorHub:  sensorHub,
		isActive:   true,
		startTime:  time.Now(),
	}, nil
}

// Shutdown gracefully stops all subsystems
func (s *System) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.isActive = false
	s.cancelFunc()
	
	// shutdown neural network
	s.neuralNet.Shutdown()
	
	// shutdown sensor systems
	s.sensorHub.Shutdown()
}

// IsActive checks if system is still running
func (s *System) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isActive
}

// GetUptime returns how long system has been running
func (s *System) GetUptime() time.Duration {
	return time.Since(s.startTime)
} 