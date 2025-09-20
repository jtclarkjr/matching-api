package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ParseAndValidateJSON parses JSON request body and validates it
func ParseAndValidateJSON(r *http.Request, dest interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	if err := validate.Struct(dest); err != nil {
		return formatValidationError(err)
	}

	return nil
}

// ValidateStruct validates a struct using validator tags
func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		return formatValidationError(err)
	}
	return nil
}

// formatValidationError formats validation errors into readable messages
func formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, e := range validationErrors {
			messages = append(messages, formatFieldError(e))
		}
		return fmt.Errorf("%s", strings.Join(messages, "; "))
	}
	return err
}

// formatFieldError formats a single field validation error
func formatFieldError(e validator.FieldError) string {
	field := e.Field()
	
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	case "dive":
		return fmt.Sprintf("%s contains invalid values", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// GetQueryParam gets a query parameter with a default value
func GetQueryParam(r *http.Request, key, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetQueryParamInt gets an integer query parameter with a default value
func GetQueryParamInt(r *http.Request, key string, defaultValue int) (int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue, nil
	}
	
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue, fmt.Errorf("invalid integer value for %s", key)
	}
	
	return intValue, nil
}

// GetQueryParamFloat gets a float query parameter with a default value
func GetQueryParamFloat(r *http.Request, key string, defaultValue float64) (float64, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue, nil
	}
	
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue, fmt.Errorf("invalid float value for %s", key)
	}
	
	return floatValue, nil
}

// IsValidGender checks if gender is valid
func IsValidGender(gender string) bool {
	validGenders := []string{"male", "female", "non-binary"}
	for _, valid := range validGenders {
		if gender == valid {
			return true
		}
	}
	return false
}

// UpdateStructFields updates non-nil fields from source to destination
func UpdateStructFields(source, dest interface{}) {
	sourceVal := reflect.ValueOf(source)
	destVal := reflect.ValueOf(dest)
	
	if sourceVal.Kind() == reflect.Ptr {
		sourceVal = sourceVal.Elem()
	}
	if destVal.Kind() == reflect.Ptr {
		destVal = destVal.Elem()
	}
	
	sourceType := sourceVal.Type()
	
	for i := 0; i < sourceVal.NumField(); i++ {
		sourceField := sourceVal.Field(i)
		sourceFieldType := sourceType.Field(i)
		destField := destVal.FieldByName(sourceFieldType.Name)
		
		if !destField.IsValid() || !destField.CanSet() {
			continue
		}
		
		// Handle pointer fields (for partial updates)
		if sourceField.Kind() == reflect.Ptr && !sourceField.IsNil() {
			destField.Set(sourceField.Elem())
		} else if sourceField.Kind() != reflect.Ptr && sourceField.IsValid() {
			destField.Set(sourceField)
		}
	}
}