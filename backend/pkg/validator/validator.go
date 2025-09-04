package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validators
	validate.RegisterValidation("decimal_positive", validateDecimalPositive)
	validate.RegisterValidation("decimal_non_negative", validateDecimalNonNegative)
	validate.RegisterValidation("asset_type", validateAssetType)
	validate.RegisterValidation("transaction_type", validateTransactionType)
	validate.RegisterValidation("alert_severity", validateAlertSeverity)
	validate.RegisterValidation("risk_status", validateRiskStatus)
}

// Validate validates a struct using the validator package
func Validate(s interface{}) error {
	return validate.Struct(s)
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
	return validate.Var(field, tag)
}

// GetValidationErrors returns formatted validation errors
func GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			errors[strings.ToLower(fieldError.Field())] = getErrorMessage(fieldError)
		}
	}

	return errors
}

func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fe.Field(), fe.Param())
	case "decimal_positive":
		return fmt.Sprintf("%s must be positive", fe.Field())
	case "decimal_non_negative":
		return fmt.Sprintf("%s must be non-negative", fe.Field())
	case "asset_type":
		return "Invalid asset type. Must be one of: STOCK, BOND, CRYPTO, COMMODITY, CASH"
	case "transaction_type":
		return "Invalid transaction type. Must be one of: BUY, SELL"
	case "alert_severity":
		return "Invalid alert severity. Must be one of: LOW, MEDIUM, HIGH, CRITICAL"
	case "risk_status":
		return "Invalid risk status. Must be one of: SAFE, WARNING, CRITICAL"
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

// Custom validators

func validateDecimalPositive(fl validator.FieldLevel) bool {
	if dec, ok := fl.Field().Interface().(decimal.Decimal); ok {
		return dec.GreaterThan(decimal.Zero)
	}
	return false
}

func validateDecimalNonNegative(fl validator.FieldLevel) bool {
	if dec, ok := fl.Field().Interface().(decimal.Decimal); ok {
		return dec.GreaterThanOrEqual(decimal.Zero)
	}
	return false
}

func validateAssetType(fl validator.FieldLevel) bool {
	assetType := fl.Field().String()
	validTypes := []string{"STOCK", "BOND", "CRYPTO", "COMMODITY", "CASH"}

	for _, validType := range validTypes {
		if assetType == validType {
			return true
		}
	}
	return false
}

func validateTransactionType(fl validator.FieldLevel) bool {
	transactionType := fl.Field().String()
	validTypes := []string{"BUY", "SELL"}

	for _, validType := range validTypes {
		if transactionType == validType {
			return true
		}
	}
	return false
}

func validateAlertSeverity(fl validator.FieldLevel) bool {
	severity := fl.Field().String()
	validSeverities := []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"}

	for _, validSeverity := range validSeverities {
		if severity == validSeverity {
			return true
		}
	}
	return false
}

func validateRiskStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := []string{"SAFE", "WARNING", "CRITICAL"}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// Email validation
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// Password strength validation
func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

	return hasUpper && hasLower && hasDigit && hasSpecial
}
