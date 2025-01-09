package nlp

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

// CommandType represents different types of commands
type CommandType string

const (
	CmdMove     CommandType = "move"
	CmdStop     CommandType = "stop"
	CmdAdjust   CommandType = "adjust"
	CmdStatus   CommandType = "status"
	CmdUnknown  CommandType = "unknown"
)

// Command represents parsed user command
type Command struct {
	Type       CommandType
	Parameters map[string]interface{}
	Priority   int
	Timestamp  time.Time
}

// Response represents system's reply
type Response struct {
	Text       string
	Sentiment  float64  // -1.0 to 1.0
	Confidence float64
	Timestamp  time.Time
}

// Processor handles natural language processing
type Processor struct {
	mu sync.RWMutex
	
	// Command processing
	commandHistory []Command
	lastCommand    *Command
	
	// Response generation
	responseHistory []Response
	lastResponse    *Response
	
	// Context management
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewProcessor creates new NLP processor
func NewProcessor() (*Processor, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Processor{
		commandHistory:  make([]Command, 0),
		responseHistory: make([]Response, 0),
		ctx:            ctx,
		cancelFunc:     cancel,
	}, nil
}

// ProcessCommand handles incoming command text
func (p *Processor) ProcessCommand(text string) (*Command, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Basic command parsing
	words := strings.Fields(strings.ToLower(text))
	if len(words) == 0 {
		return nil, errors.New("empty command")
	}
	
	cmd := &Command{
		Type:       p.determineCommandType(words),
		Parameters: make(map[string]interface{}),
		Priority:   1,
		Timestamp:  time.Now(),
	}
	
	// Parse parameters based on command type
	switch cmd.Type {
	case CmdMove:
		p.parseMovementParams(words, cmd)
	case CmdAdjust:
		p.parseAdjustmentParams(words, cmd)
	case CmdStatus:
		// No parameters needed
	case CmdStop:
		cmd.Priority = 10 // High priority for stop command
	}
	
	// Store command in history
	p.commandHistory = append(p.commandHistory, *cmd)
	if len(p.commandHistory) > 1000 {
		p.commandHistory = p.commandHistory[1:]
	}
	p.lastCommand = cmd
	
	return cmd, nil
}

// determineCommandType identifies command type from words
func (p *Processor) determineCommandType(words []string) CommandType {
	if len(words) == 0 {
		return CmdUnknown
	}
	
	// Simple keyword matching
	moveKeywords := []string{"move", "go", "rotate", "turn"}
	stopKeywords := []string{"stop", "halt", "freeze"}
	adjustKeywords := []string{"adjust", "change", "modify"}
	statusKeywords := []string{"status", "state", "condition"}
	
	for _, word := range words {
		if containsWord(moveKeywords, word) {
			return CmdMove
		}
		if containsWord(stopKeywords, word) {
			return CmdStop
		}
		if containsWord(adjustKeywords, word) {
			return CmdAdjust
		}
		if containsWord(statusKeywords, word) {
			return CmdStatus
		}
	}
	
	return CmdUnknown
}

// parseMovementParams extracts movement parameters
func (p *Processor) parseMovementParams(words []string, cmd *Command) {
	for i := 0; i < len(words)-1; i++ {
		switch words[i] {
		case "speed":
			if speed, ok := parseFloat(words[i+1]); ok {
				cmd.Parameters["speed"] = speed
			}
		case "direction":
			cmd.Parameters["direction"] = words[i+1]
		case "distance":
			if dist, ok := parseFloat(words[i+1]); ok {
				cmd.Parameters["distance"] = dist
			}
		}
	}
}

// parseAdjustmentParams extracts adjustment parameters
func (p *Processor) parseAdjustmentParams(words []string, cmd *Command) {
	for i := 0; i < len(words)-1; i++ {
		switch words[i] {
		case "intensity":
			if intensity, ok := parseFloat(words[i+1]); ok {
				cmd.Parameters["intensity"] = intensity
			}
		case "sensitivity":
			if sensitivity, ok := parseFloat(words[i+1]); ok {
				cmd.Parameters["sensitivity"] = sensitivity
			}
		}
	}
}

// GenerateResponse creates appropriate response
func (p *Processor) GenerateResponse(cmd *Command) (*Response, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	response := &Response{
		Confidence: 0.8,
		Timestamp:  time.Now(),
	}
	
	// Generate response based on command type
	switch cmd.Type {
	case CmdMove:
		response.Text = "Moving as requested, tovarisch"
		response.Sentiment = 0.5
	case CmdStop:
		response.Text = "Emergency stop initiated! Bozhe moy!"
		response.Sentiment = -0.3
		response.Confidence = 1.0
	case CmdAdjust:
		response.Text = "Adjusting parameters, one moment please"
		response.Sentiment = 0.2
	case CmdStatus:
		response.Text = "All systems operational, running like Kalashnikov"
		response.Sentiment = 0.8
	default:
		response.Text = "Command not understood, try again comrade"
		response.Sentiment = -0.1
		response.Confidence = 0.4
	}
	
	// Store response in history
	p.responseHistory = append(p.responseHistory, *response)
	if len(p.responseHistory) > 1000 {
		p.responseHistory = p.responseHistory[1:]
	}
	p.lastResponse = response
	
	return response, nil
}

// GetLastCommand returns most recent command
func (p *Processor) GetLastCommand() *Command {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastCommand
}

// GetLastResponse returns most recent response
func (p *Processor) GetLastResponse() *Response {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastResponse
}

// Shutdown stops NLP processor
func (p *Processor) Shutdown() {
	p.cancelFunc()
}

// Helper functions

func containsWord(words []string, target string) bool {
	for _, word := range words {
		if word == target {
			return true
		}
	}
	return false
}

func parseFloat(s string) (float64, bool) {
	// TODO: implement proper float parsing
	return 0.0, false
} 