package motion

import (
	"errors"
	"math"
	"sync"
	"time"
)

// MotorID represents unique identifier for each motor
type MotorID string

// MotorType defines different types of motors
type MotorType int

const (
	MotorServo MotorType = iota
	MotorStepper
	MotorDC
)

// Motor represents single motor unit
type Motor struct {
	ID          MotorID
	Type        MotorType
	Position    float64  // current position in degrees
	Speed       float64  // current speed in degrees/second
	MaxSpeed    float64  // maximum allowed speed
	MinPosition float64  // minimum allowed position
	MaxPosition float64  // maximum allowed position
	IsEnabled   bool
}

// Controller manages all motion systems
type Controller struct {
	mu      sync.RWMutex
	motors  map[MotorID]*Motor
	running bool
	
	// Movement patterns
	patterns map[string]MovementPattern
	
	// Control channels
	controlChan chan MotorCommand
	done        chan struct{}
}

// MotorCommand represents command for motor
type MotorCommand struct {
	ID       MotorID
	Position float64
	Speed    float64
}

// MovementPattern represents predefined movement sequence
type MovementPattern struct {
	Name     string
	Commands []MotorCommand
	Duration time.Duration
}

// NewController initializes motion control system
func NewController() (*Controller, error) {
	c := &Controller{
		motors:      make(map[MotorID]*Motor),
		patterns:    make(map[string]MovementPattern),
		controlChan: make(chan MotorCommand, 100),
		done:        make(chan struct{}),
		running:     true,
	}
	
	// Initialize default motors
	defaultMotors := []Motor{
		{
			ID:          "servo_1",
			Type:        MotorServo,
			MaxSpeed:    180.0,
			MinPosition: 0.0,
			MaxPosition: 180.0,
			IsEnabled:   true,
		},
		{
			ID:          "servo_2",
			Type:        MotorServo,
			MaxSpeed:    180.0,
			MinPosition: 0.0,
			MaxPosition: 180.0,
			IsEnabled:   true,
		},
		// Add more motors as needed
	}
	
	for _, m := range defaultMotors {
		motor := m // Create new variable to avoid pointer issues
		c.motors[motor.ID] = &motor
	}
	
	go c.processCommands()
	
	return c, nil
}

// processCommands handles incoming motor commands
func (c *Controller) processCommands() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case cmd := <-c.controlChan:
			c.executeCommand(cmd)
		case <-c.done:
			return
		case <-ticker.C:
			c.updateMotorStates()
		}
	}
}

// executeCommand processes single motor command
func (c *Controller) executeCommand(cmd MotorCommand) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	motor, exists := c.motors[cmd.ID]
	if !exists {
		return errors.New("motor not found")
	}
	
	if !motor.IsEnabled {
		return errors.New("motor is disabled")
	}
	
	// Validate position
	if cmd.Position < motor.MinPosition || cmd.Position > motor.MaxPosition {
		return errors.New("position out of range")
	}
	
	// Validate speed
	speed := math.Abs(cmd.Speed)
	if speed > motor.MaxSpeed {
		speed = motor.MaxSpeed
	}
	
	motor.Position = cmd.Position
	motor.Speed = speed
	
	return nil
}

// updateMotorStates updates all motor positions based on current speeds
func (c *Controller) updateMotorStates() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for _, motor := range c.motors {
		if !motor.IsEnabled {
			continue
		}
		
		// Update position based on speed
		// This is simplified; real implementation would be more complex
		delta := motor.Speed * 0.01 // 10ms tick
		newPos := motor.Position + delta
		
		// Ensure position stays within bounds
		if newPos < motor.MinPosition {
			newPos = motor.MinPosition
			motor.Speed = 0
		} else if newPos > motor.MaxPosition {
			newPos = motor.MaxPosition
			motor.Speed = 0
		}
		
		motor.Position = newPos
	}
}

// AddPattern adds new movement pattern
func (c *Controller) AddPattern(pattern MovementPattern) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.patterns[pattern.Name] = pattern
}

// ExecutePattern runs predefined movement pattern
func (c *Controller) ExecutePattern(name string) error {
	c.mu.RLock()
	pattern, exists := c.patterns[name]
	c.mu.RUnlock()
	
	if !exists {
		return errors.New("pattern not found")
	}
	
	go func() {
		for _, cmd := range pattern.Commands {
			if !c.running {
				return
			}
			c.controlChan <- cmd
			time.Sleep(pattern.Duration / time.Duration(len(pattern.Commands)))
		}
	}()
	
	return nil
}

// Shutdown stops motion control system
func (c *Controller) Shutdown() {
	c.mu.Lock()
	c.running = false
	c.mu.Unlock()
	
	close(c.done)
	close(c.controlChan)
	
	// Disable all motors
	for _, motor := range c.motors {
		motor.IsEnabled = false
		motor.Speed = 0
	}
} 