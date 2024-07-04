package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

// Task represents a task in the database
type Task struct {
	TaskID        string                 `gorm:"primaryKey" json:"task_id"`          // Task identifier
	DateTime      string                 `json:"date_time"`                          // Task creation time
	Company       string                 `json:"company"`                            // Company name
	Args          map[string]interface{} `gorm:"type:json" json:"args"`              // Task arguments as JSON
	Status        string                 `json:"status"`                             // Task status
	SuccessJobIDs []string               `gorm:"type:text[]" json:"success_job_ids"` // IDs of successful jobs
	FailedJobIDs  []string               `gorm:"type:text[]" json:"failed_job_ids"`  // IDs of failed jobs
	SuccessRate   float64                `json:"success_rate"`                       // Success rate
}

// BeforeSave is a GORM hook that runs before saving a Task to the database
func (t *Task) BeforeSave(tx *gorm.DB) (err error) {
	// Marshal Args to JSON string
	argsJSON, err := json.Marshal(t.Args)
	if err != nil {
		return err
	}
	tx.Statement.SetColumn("Args", string(argsJSON))
	return nil
}

// AfterFind is a GORM hook that runs after finding a Task in the database
func (t *Task) AfterFind(tx *gorm.DB) (err error) {
	// Unmarshal JSON string to Args map
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(t.Args), &args); err != nil {
		return err
	}
	t.Args = args
	return nil
}
