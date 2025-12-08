package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func DecodeAndValidateJSON(body io.ReadCloser, v interface{}) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {

		if strings.Contains(err.Error(), "unknown field") {
			field := extractUnknownField(err.Error())
			return fmt.Errorf("unknown field '%s' in request body", field)
		}

		if strings.Contains(err.Error(), "cannot unmarshal") {
			return fmt.Errorf("invalid value type: %s", err.Error())
		}

		return fmt.Errorf("invalid JSON format: %v", err)
	}

	return nil
}

func extractUnknownField(msg string) string {
	parts := strings.Split(msg, "\"")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}
