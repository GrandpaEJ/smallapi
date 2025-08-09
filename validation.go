package smallapi

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Validator provides data validation functionality
type Validator struct {
	rules map[string][]ValidationRule
}

// ValidationRule represents a single validation rule
type ValidationRule struct {
	Type    string
	Value   string
	Message string
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string][]ValidationRule),
	}
}

// Validate validates a struct based on field tags
func (c *Context) Validate(v interface{}) error {
	return validateStruct(v)
}

// validateStruct validates a struct using reflection and tags
func validateStruct(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return errors.New("validation target must be a struct")
	}
	
	typ := val.Type()
	var errs []string
	
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		
		// Get validation tag
		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}
		
		// Parse validation rules
		rules := parseValidationTag(tag)
		
		// Validate field
		if err := validateField(field, fieldType.Name, rules); err != nil {
			errs = append(errs, err.Error())
		}
	}
	
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	
	return nil
}

// parseValidationTag parses a validation tag into rules
func parseValidationTag(tag string) []ValidationRule {
	var rules []ValidationRule
	
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		// Split rule and value (e.g., "min=5")
		ruleParts := strings.SplitN(part, "=", 2)
		ruleType := ruleParts[0]
		ruleValue := ""
		if len(ruleParts) > 1 {
			ruleValue = ruleParts[1]
		}
		
		rules = append(rules, ValidationRule{
			Type:  ruleType,
			Value: ruleValue,
		})
	}
	
	return rules
}

// validateField validates a single field against rules
func validateField(field reflect.Value, fieldName string, rules []ValidationRule) error {
	for _, rule := range rules {
		if err := applyValidationRule(field, fieldName, rule); err != nil {
			return err
		}
	}
	return nil
}

// applyValidationRule applies a single validation rule
func applyValidationRule(field reflect.Value, fieldName string, rule ValidationRule) error {
	switch rule.Type {
	case "required":
		return validateRequired(field, fieldName)
	case "min":
		return validateMin(field, fieldName, rule.Value)
	case "max":
		return validateMax(field, fieldName, rule.Value)
	case "email":
		return validateEmail(field, fieldName)
	case "url":
		return validateURL(field, fieldName)
	case "numeric":
		return validateNumeric(field, fieldName)
	case "alpha":
		return validateAlpha(field, fieldName)
	case "alphanum":
		return validateAlphaNum(field, fieldName)
	case "regex":
		return validateRegex(field, fieldName, rule.Value)
	default:
		return fmt.Errorf("unknown validation rule: %s", rule.Type)
	}
}

// validateRequired checks if field is not empty
func validateRequired(field reflect.Value, fieldName string) error {
	switch field.Kind() {
	case reflect.String:
		if field.String() == "" {
			return fmt.Errorf("%s is required", fieldName)
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if field.Len() == 0 {
			return fmt.Errorf("%s is required", fieldName)
		}
	case reflect.Ptr, reflect.Interface:
		if field.IsNil() {
			return fmt.Errorf("%s is required", fieldName)
		}
	}
	return nil
}

// validateMin checks minimum length or value
func validateMin(field reflect.Value, fieldName, minStr string) error {
	min, err := strconv.Atoi(minStr)
	if err != nil {
		return fmt.Errorf("invalid min value: %s", minStr)
	}
	
	switch field.Kind() {
	case reflect.String:
		if len(field.String()) < min {
			return fmt.Errorf("%s must be at least %d characters", fieldName, min)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() < int64(min) {
			return fmt.Errorf("%s must be at least %d", fieldName, min)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() < float64(min) {
			return fmt.Errorf("%s must be at least %d", fieldName, min)
		}
	case reflect.Slice, reflect.Array:
		if field.Len() < min {
			return fmt.Errorf("%s must have at least %d items", fieldName, min)
		}
	}
	return nil
}

// validateMax checks maximum length or value
func validateMax(field reflect.Value, fieldName, maxStr string) error {
	max, err := strconv.Atoi(maxStr)
	if err != nil {
		return fmt.Errorf("invalid max value: %s", maxStr)
	}
	
	switch field.Kind() {
	case reflect.String:
		if len(field.String()) > max {
			return fmt.Errorf("%s must be at most %d characters", fieldName, max)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() > int64(max) {
			return fmt.Errorf("%s must be at most %d", fieldName, max)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() > float64(max) {
			return fmt.Errorf("%s must be at most %d", fieldName, max)
		}
	case reflect.Slice, reflect.Array:
		if field.Len() > max {
			return fmt.Errorf("%s must have at most %d items", fieldName, max)
		}
	}
	return nil
}

// validateEmail checks if field is a valid email
func validateEmail(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("%s must be a string for email validation", fieldName)
	}
	
	email := field.String()
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("%s must be a valid email address", fieldName)
	}
	
	return nil
}

// validateURL checks if field is a valid URL
func validateURL(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("%s must be a string for URL validation", fieldName)
	}
	
	url := field.String()
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	
	if !urlRegex.MatchString(url) {
		return fmt.Errorf("%s must be a valid URL", fieldName)
	}
	
	return nil
}

// validateNumeric checks if field contains only numbers
func validateNumeric(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("%s must be a string for numeric validation", fieldName)
	}
	
	value := field.String()
	numericRegex := regexp.MustCompile(`^[0-9]+$`)
	
	if !numericRegex.MatchString(value) {
		return fmt.Errorf("%s must contain only numbers", fieldName)
	}
	
	return nil
}

// validateAlpha checks if field contains only letters
func validateAlpha(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("%s must be a string for alpha validation", fieldName)
	}
	
	value := field.String()
	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	
	if !alphaRegex.MatchString(value) {
		return fmt.Errorf("%s must contain only letters", fieldName)
	}
	
	return nil
}

// validateAlphaNum checks if field contains only letters and numbers
func validateAlphaNum(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("%s must be a string for alphanumeric validation", fieldName)
	}
	
	value := field.String()
	alphaNumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	
	if !alphaNumRegex.MatchString(value) {
		return fmt.Errorf("%s must contain only letters and numbers", fieldName)
	}
	
	return nil
}

// validateRegex checks if field matches a custom regex pattern
func validateRegex(field reflect.Value, fieldName, pattern string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("%s must be a string for regex validation", fieldName)
	}
	
	value := field.String()
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %s", pattern)
	}
	
	if !regex.MatchString(value) {
		return fmt.Errorf("%s does not match required pattern", fieldName)
	}
	
	return nil
}
