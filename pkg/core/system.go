package core

import (
	"context"
	"sync"
	"time"

	"github.com/sashalind/sex-artifical-intelligence/pkg/behavior"
	"github.com/sashalind/sex-artifical-intelligence/pkg/motion"
	"github.com/sashalind/sex-artifical-intelligence/pkg/neural"
	"github.com/sashalind/sex-artifical-intelligence/pkg/nlp"
	"github.com/sashalind/sex-artifical-intelligence/pkg/sensor"
)

// System represents main control system blyat
type System struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	
	neuralNet  *neural.Network
	sensorHub  *sensor.Hub
	motionCtrl *motion.Controller
	behavior   *behavior.Analyzer
	nlpProc    *nlp.Processor
	
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
	
	motionCtrl, err := motion.NewController()
	if err != nil {
		cancel()
		return nil, err
	}
	
	behaviorAnalyzer, err := behavior.NewAnalyzer()
	if err != nil {
		cancel()
		return nil, err
	}
	
	nlpProcessor, err := nlp.NewProcessor()
	if err != nil {
		cancel()
		return nil, err
	}
	
	sys := &System{
		ctx:        ctx,
		cancelFunc: cancel,
		neuralNet:  neuralNet,
		sensorHub:  sensorHub,
		motionCtrl: motionCtrl,
		behavior:   behaviorAnalyzer,
		nlpProc:    nlpProcessor,
		isActive:   true,
		startTime:  time.Now(),
	}
	
	// Start behavior analysis based on sensor data
	go sys.analyzeBehavior()
	
	return sys, nil
}

// ProcessCommand handles user command
func (s *System) ProcessCommand(text string) (*nlp.Response, error) {
	// Parse command using NLP
	cmd, err := s.nlpProc.ProcessCommand(text)
	if err != nil {
		return nil, err
	}
	
	// Handle command based on type
	switch cmd.Type {
	case nlp.CmdMove:
		if err := s.handleMovement(cmd); err != nil {
			return nil, err
		}
	case nlp.CmdStop:
		if err := s.handleStop(cmd); err != nil {
			return nil, err
		}
	case nlp.CmdAdjust:
		if err := s.handleAdjustment(cmd); err != nil {
			return nil, err
		}
	}
	
	// Generate response
	return s.nlpProc.GenerateResponse(cmd)
}

// Command handlers

func (s *System) handleMovement(cmd *nlp.Command) error {
	// Extract movement parameters
	speed, ok := cmd.Parameters["speed"].(float64)
	if !ok {
		speed = 1.0 // default speed
	}
	
	// Create motor command
	motorCmd := motion.MotorCommand{
		ID:       "servo_1", // TODO: determine appropriate motor
		Speed:    speed,
		Position: 90.0, // TODO: calculate from direction
	}
	
	// Send command to motion controller
	return s.motionCtrl.ExecuteCommand(motorCmd)
}

func (s *System) handleStop(cmd *nlp.Command) error {
	// Stop all motors
	for _, motor := range s.motionCtrl.GetMotors() {
		stopCmd := motion.MotorCommand{
			ID:       motor.ID,
			Speed:    0,
			Position: motor.Position,
		}
		if err := s.motionCtrl.ExecuteCommand(stopCmd); err != nil {
			return err
		}
	}
	return nil
}

func (s *System) handleAdjustment(cmd *nlp.Command) error {
	// TODO: implement parameter adjustment
	return nil
}

// analyzeBehavior processes sensor data for behavioral patterns
func (s *System) analyzeBehavior() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if !s.isActive {
				return
			}
			
			// Get latest sensor data
			touchData := s.sensorHub.GetSensorData(sensor.TypeTouch)
			pressureData := s.sensorHub.GetSensorData(sensor.TypePressure)
			motionData := s.sensorHub.GetSensorData(sensor.TypeMotion)
			
			if len(touchData) == 0 || len(pressureData) == 0 || len(motionData) == 0 {
				continue
			}
			
			// Calculate behavior metrics
			metrics := behavior.PatternMetrics{
				Intensity:    calculateIntensity(touchData, pressureData),
				Frequency:    calculateFrequency(motionData),
				Duration:     1.0, // TODO: implement duration calculation
				Consistency: calculateConsistency(touchData, pressureData, motionData),
			}
			
			// Send metrics for analysis
			s.behavior.AddMetrics(metrics)
		}
	}
}

// Helper functions for behavior analysis

func calculateIntensity(touch, pressure []float64) float64 {
	if len(touch) == 0 || len(pressure) == 0 {
		return 0.0
	}
	
	// Use latest readings
	touchIntensity := touch[len(touch)-1]
	pressureIntensity := pressure[len(pressure)-1]
	
	// Normalize and combine
	return (touchIntensity + pressureIntensity) / 2.0
}

func calculateFrequency(motion []float64) float64 {
	if len(motion) < 2 {
		return 0.0
	}
	
	// Calculate rate of change in motion
	var changes float64
	for i := 1; i < len(motion); i++ {
		if motion[i] != motion[i-1] {
			changes++
		}
	}
	
	return changes / float64(len(motion))
}

func calculateConsistency(touch, pressure, motion []float64) float64 {
	// Simple variance-based consistency measure
	allData := append(append(touch, pressure...), motion...)
	if len(allData) < 2 {
		return 1.0
	}
	
	var mean, variance float64
	for _, v := range allData {
		mean += v
	}
	mean /= float64(len(allData))
	
	for _, v := range allData {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(allData))
	
	// Convert variance to consistency score (0-1)
	consistency := 1.0 / (1.0 + variance)
	if consistency > 1.0 {
		consistency = 1.0
	}
	
	return consistency
}

// Shutdown gracefully stops all subsystems
func (s *System) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.isActive = false
	s.cancelFunc()
	
	// shutdown all subsystems
	s.neuralNet.Shutdown()
	s.sensorHub.Shutdown()
	s.motionCtrl.Shutdown()
	s.behavior.Shutdown()
	s.nlpProc.Shutdown()
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