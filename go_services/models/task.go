package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Task struct {
	TaskID         string   `gorm:"primaryKey" json:"task_id"`          // Task identifier
	ContainerID    string   `json:"container_id"`                       // Docker Container ID
	DateTime       string   `json:"date_time"`                          // Task creation time
	Args           JSONMap  `gorm:"type:json" json:"args"`              // Task arguments as JSON
	Status         string   `json:"status"`                             // Task status
	SuccessJobIDs  []string `gorm:"type:text[]" json:"success_job_ids"` // IDs of successful jobs
	FailedJobIDs   []string `gorm:"type:text[]" json:"failed_job_ids"`  // IDs of failed jobs
	CompletionRate float64  `json:"completion_rate"`                    // Completion rate
	IsRetryTask    bool     `json:"is_retry"`                           // Task is retry
	ParentTaskID   string   `json:"parent_task_id"`                     // Only if task is retried
}

// JSONMap is a custom type to handle JSON encoding/decoding
type JSONMap map[string]interface{}

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
