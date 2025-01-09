package safety

import (
	"log"
	"sync"
	"time"

	"github.com/sashalind/sex-artifical-intelligence/pkg/core"
)

// SafetyLevel represents system safety status
type SafetyLevel int

const (
	SafetyNormal SafetyLevel = iota
	SafetyWarning
	SafetyCritical
	SafetyEmergency
)

// SafetyMonitor handles system safety
type SafetyMonitor struct {
	system     *core.System
	mu         sync.RWMutex
	
	// safety parameters
	currentLevel SafetyLevel
	lastCheck    time.Time
	warnings     []string
}

var monitor *SafetyMonitor

// InitializeSafetyProtocols sets up safety systems
func InitializeSafetyProtocols(sys *core.System) {
	monitor = &SafetyMonitor{
		system:      sys,
		currentLevel: SafetyNormal,
		lastCheck:    time.Now(),
		warnings:     make([]string, 0),
	}
	
	go monitor.runSafetyChecks()
}

// runSafetyChecks performs periodic system safety verification
func (s *SafetyMonitor) runSafetyChecks() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		if !s.system.IsActive() {
			return
		}
		
		s.performSafetyCheck()
	}
}

// performSafetyCheck runs single safety verification
func (s *SafetyMonitor) performSafetyCheck() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.lastCheck = time.Now()
	
	// TODO: implement actual safety checks
	// For now just log that we're checking
	log.Printf("Safety check performed at %v - Status: %v\n", 
		s.lastCheck.Format(time.RFC3339),
		s.currentLevel)
}

// AddWarning adds new safety warning
func (s *SafetyMonitor) AddWarning(warning string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.warnings = append(s.warnings, warning)
	
	if len(s.warnings) > 10 {
		s.currentLevel = SafetyWarning
	}
	
	if len(s.warnings) > 20 {
		s.currentLevel = SafetyCritical
	}
}

// GetCurrentLevel returns current safety level
func (s *SafetyMonitor) GetCurrentLevel() SafetyLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentLevel
}

// GetWarnings returns all active warnings
func (s *SafetyMonitor) GetWarnings() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string{}, s.warnings...) 