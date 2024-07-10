package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// Value implements the driver.Valuer interface for TaskStatus
func (ts TaskStatus) Value() (driver.Value, error) {
	return ts.String(), nil
}

// Scan implements the sql.Scanner interface for TaskStatus
func (ts *TaskStatus) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.New("invalid type assertion")
	}

	switch str {
	case "started":
		*ts = Started
	case "in_progress":
		*ts = InProgress
	case "error":
		*ts = Error
	case "completed":
		*ts = Completed
	default:
		return fmt.Errorf("invalid TaskStatus: %s", str)
	}

	return nil
}

func (j JSONMap) Value() (driver.Value, error) {
	value, err := json.Marshal(j)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSONMap: %w", err)
	}
	return string(value), nil
}

func (j *JSONMap) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		if err := json.Unmarshal(v, j); err != nil {
			return fmt.Errorf("error unmarshaling JSONMap: %w", err)
		}
	case string:
		if err := json.Unmarshal([]byte(v), j); err != nil {
			return fmt.Errorf("error unmarshaling JSONMap: %w", err)
		}
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
	return nil
}
