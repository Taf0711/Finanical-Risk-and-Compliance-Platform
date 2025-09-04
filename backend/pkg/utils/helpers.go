package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// GenerateID generates a random hexadecimal ID
func GenerateID(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// FormatCurrency formats a decimal value as currency
func FormatCurrency(amount decimal.Decimal) string {
	return fmt.Sprintf("$%s", amount.StringFixed(2))
}

// FormatPercentage formats a decimal value as percentage
func FormatPercentage(value decimal.Decimal) string {
	return fmt.Sprintf("%.2f%%", value.InexactFloat64()*100)
}

// ParseDecimal safely parses a string to decimal
func ParseDecimal(s string) (decimal.Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(s)
}

// RoundToDecimalPlaces rounds a float64 to specified decimal places
func RoundToDecimalPlaces(value float64, places int) float64 {
	multiplier := math.Pow(10, float64(places))
	return math.Round(value*multiplier) / multiplier
}

// StringToFloat64 safely converts string to float64
func StringToFloat64(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

// StringToInt safely converts string to int
func StringToInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// CalculatePercentageChange calculates percentage change between two values
func CalculatePercentageChange(oldValue, newValue decimal.Decimal) decimal.Decimal {
	if oldValue.IsZero() {
		return decimal.Zero
	}
	return newValue.Sub(oldValue).Div(oldValue).Mul(decimal.NewFromInt(100))
}

// ClampFloat64 clamps a float64 value between min and max
func ClampFloat64(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Contains checks if a slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from a slice
func RemoveDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
