package rules

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

type KYCAMLChecker struct {
	SuspiciousAmountThreshold decimal.Decimal // e.g., $10,000
	HighRiskCountries         []string
	VelocityTimeWindow        time.Duration // e.g., 24 hours
	VelocityCountThreshold    int           // Max transactions in time window
}

func NewKYCAMLChecker() *KYCAMLChecker {
	return &KYCAMLChecker{
		SuspiciousAmountThreshold: decimal.NewFromInt(10000),
		HighRiskCountries: []string{
			"North Korea", "Iran", "Syria", "Cuba", "Venezuela",
		},
		VelocityTimeWindow:     24 * time.Hour,
		VelocityCountThreshold: 10,
	}
}

// CheckTransaction performs KYC/AML checks on a transaction
func (k *KYCAMLChecker) CheckTransaction(tx *models.Transaction, recentTransactions []models.Transaction) AMLCheckResult {
	result := AMLCheckResult{
		TransactionID: tx.ID,
		Passed:        true,
		RiskScore:     0,
		Flags:         []string{},
	}

	// Check 1: Large transaction amount
	if tx.Amount.GreaterThan(k.SuspiciousAmountThreshold) {
		result.Flags = append(result.Flags, "LARGE_TRANSACTION")
		result.RiskScore += 30
	}

	// Check 2: Velocity check (too many transactions)
	velocityCount := k.countRecentTransactions(recentTransactions, k.VelocityTimeWindow)
	if velocityCount > k.VelocityCountThreshold {
		result.Flags = append(result.Flags, "HIGH_VELOCITY")
		result.RiskScore += 40
	}

	// Check 3: Structuring detection (multiple transactions just below threshold)
	if k.detectStructuring(recentTransactions) {
		result.Flags = append(result.Flags, "POSSIBLE_STRUCTURING")
		result.RiskScore += 50
	}

	// Check 4: Round amount detection
	if k.isRoundAmount(tx.Amount) {
		result.Flags = append(result.Flags, "ROUND_AMOUNT")
		result.RiskScore += 10
	}

	// Determine if transaction should be flagged
	if result.RiskScore >= 50 {
		result.Passed = false
		result.RequiresReview = true
	}

	return result
}

func (k *KYCAMLChecker) countRecentTransactions(transactions []models.Transaction, window time.Duration) int {
	cutoff := time.Now().Add(-window)
	count := 0

	for _, tx := range transactions {
		if tx.CreatedAt.After(cutoff) {
			count++
		}
	}

	return count
}

func (k *KYCAMLChecker) detectStructuring(transactions []models.Transaction) bool {
	// Look for multiple transactions just below the threshold
	threshold90Percent := k.SuspiciousAmountThreshold.Mul(decimal.NewFromFloat(0.9))
	suspiciousCount := 0

	cutoff := time.Now().Add(-24 * time.Hour)

	for _, tx := range transactions {
		if tx.CreatedAt.After(cutoff) &&
			tx.Amount.GreaterThan(threshold90Percent) &&
			tx.Amount.LessThan(k.SuspiciousAmountThreshold) {
			suspiciousCount++
		}
	}

	return suspiciousCount >= 3
}

func (k *KYCAMLChecker) isRoundAmount(amount decimal.Decimal) bool {
	// Check if amount is a round number (e.g., 5000, 10000)
	amountFloat := amount.InexactFloat64()
	return amountFloat == float64(int(amountFloat)) && int(amountFloat)%1000 == 0
}

type AMLCheckResult struct {
	TransactionID  uuid.UUID
	Passed         bool
	RequiresReview bool
	RiskScore      int
	Flags          []string
}
