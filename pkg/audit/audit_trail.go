package audit

import (
	"encoding/json"
	"fmt"
)

// FormatValue formats a value for audit trail storage
func FormatValue(value interface{}) (*string, error) {
	if value == nil {
		return nil, nil
	}

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to format audit value: %w", err)
	}

	result := string(jsonBytes)
	return &result, nil
}

// CompareValues compares two values and returns the old and new values formatted for audit
func CompareValues(oldValue, newValue interface{}) (*string, *string, error) {
	oldFormatted, err := FormatValue(oldValue)
	if err != nil {
		return nil, nil, err
	}

	newFormatted, err := FormatValue(newValue)
	if err != nil {
		return nil, nil, err
	}

	return oldFormatted, newFormatted, nil
}
