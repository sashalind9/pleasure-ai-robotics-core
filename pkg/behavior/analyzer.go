package behavior

import (
	"encoding/json"
	"math"
	"sync"
	"time"
)

// BehaviorType represents different types of behaviors
type BehaviorType string

const (
	BehaviorNormal     BehaviorType = "normal"
	BehaviorAggressive BehaviorType = "aggressive"
	BehaviorPassive    BehaviorType = "passive"
	BehaviorErratic    BehaviorType = "erratic"
)

// BehaviorPattern represents detected behavior pattern
type BehaviorPattern struct {
	Type       BehaviorType     `json:"type"`
	Confidence float64         `json:"confidence"`
	Timestamp  time.Time       `json:"timestamp"`
	Metrics    PatternMetrics  `json:"metrics"`
}

// PatternMetrics contains behavioral measurements
type PatternMetrics struct {
	Intensity    float64 `json:"intensity"`
	Frequency    float64 `json:"frequency"`
	Duration     float64 `json:"duration"`
	Consistency  float64 `json:"consistency"`
}

// Analyzer processes behavioral patterns
type Analyzer struct {
	mu           sync.RWMutex
	patterns     []BehaviorPattern
	currentState BehaviorType
	
	// Analysis parameters
	threshold    float64
	windowSize   time.Duration
	
	// Channels for real-time processing
	inputChan    chan PatternMetrics
	done         chan struct{}
}

// NewAnalyzer creates new behavior analysis system
func NewAnalyzer() (*Analyzer, error) {
	a := &Analyzer{
		patterns:     make([]BehaviorPattern, 0),
		currentState: BehaviorNormal,
		threshold:    0.75,
		windowSize:   5 * time.Minute,
		inputChan:    make(chan PatternMetrics, 100),
		done:         make(chan struct{}),
	}
	
	go a.processPatterns()
	
	return a, nil
}

// processPatterns analyzes incoming behavioral data
func (a *Analyzer) processPatterns() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	var buffer []PatternMetrics
	
	for {
		select {
		case metrics := <-a.inputChan:
			buffer = append(buffer, metrics)
			if len(buffer) > 60 { // Keep last minute of data
				buffer = buffer[1:]
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				pattern := a.analyzeBuffer(buffer)
				a.addPattern(pattern)
			}
		case <-a.done:
			return
		}
	}
}

// analyzeBuffer processes collected metrics
func (a *Analyzer) analyzeBuffer(buffer []PatternMetrics) BehaviorPattern {
	if len(buffer) == 0 {
		return BehaviorPattern{
			Type:       BehaviorNormal,
			Confidence: 1.0,
			Timestamp:  time.Now(),
		}
	}
	
	// Calculate average metrics
	var avgIntensity, avgFrequency, avgDuration, avgConsistency float64
	for _, m := range buffer {
		avgIntensity += m.Intensity
		avgFrequency += m.Frequency
		avgDuration += m.Duration
		avgConsistency += m.Consistency
	}
	
	n := float64(len(buffer))
	avgIntensity /= n
	avgFrequency /= n
	avgDuration /= n
	avgConsistency /= n
	
	// Determine behavior type based on metrics
	behaviorType := a.classifyBehavior(avgIntensity, avgFrequency)
	confidence := a.calculateConfidence(avgConsistency)
	
	return BehaviorPattern{
		Type:       behaviorType,
		Confidence: confidence,
		Timestamp:  time.Now(),
		Metrics: PatternMetrics{
			Intensity:    avgIntensity,
			Frequency:    avgFrequency,
			Duration:     avgDuration,
			Consistency:  avgConsistency,
		},
	}
}

// classifyBehavior determines behavior type from metrics
func (a *Analyzer) classifyBehavior(intensity, frequency float64) BehaviorType {
	// Simple classification based on intensity and frequency
	if intensity > 0.8 && frequency > 0.8 {
		return BehaviorAggressive
	} else if intensity < 0.2 && frequency < 0.2 {
		return BehaviorPassive
	} else if math.Abs(intensity-frequency) > 0.5 {
		return BehaviorErratic
	}
	return BehaviorNormal
}

// calculateConfidence determines confidence level
func (a *Analyzer) calculateConfidence(consistency float64) float64 {
	// Simple linear confidence based on consistency
	confidence := consistency
	if confidence > 1.0 {
		confidence = 1.0
	} else if confidence < 0.0 {
		confidence = 0.0
	}
	return confidence
}

// addPattern stores new behavior pattern
func (a *Analyzer) addPattern(pattern BehaviorPattern) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.patterns = append(a.patterns, pattern)
	if len(a.patterns) > 1000 {
		a.patterns = a.patterns[1:]
	}
	
	// Update current state if confidence is high enough
	if pattern.Confidence >= a.threshold {
		a.currentState = pattern.Type
	}
}

// GetCurrentState returns current behavior state
func (a *Analyzer) GetCurrentState() BehaviorType {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentState
}

// GetPatternHistory returns recent behavior patterns
func (a *Analyzer) GetPatternHistory() []BehaviorPattern {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	// Return copy to prevent data races
	patterns := make([]BehaviorPattern, len(a.patterns))
	copy(patterns, a.patterns)
	return patterns
}

// AddMetrics adds new behavioral metrics for analysis
func (a *Analyzer) AddMetrics(metrics PatternMetrics) {
	a.inputChan <- metrics
}

// Shutdown stops behavior analysis
func (a *Analyzer) Shutdown() {
	close(a.done)
	close(a.inputChan)
} 