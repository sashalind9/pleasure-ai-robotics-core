package utils

import "math"

// ReLU implements Rectified Linear Unit activation
func ReLU(x float64) float64 {
	if x < 0 {
		return 0
	}
	return x
}

// Sigmoid implements sigmoid activation function
func Sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// Tanh implements hyperbolic tangent activation
func Tanh(x float64) float64 {
	return math.Tanh(x)
}

// LeakyReLU implements Leaky ReLU activation
func LeakyReLU(x float64) float64 {
	if x < 0 {
		return 0.01 * x
	}
	return x
} 