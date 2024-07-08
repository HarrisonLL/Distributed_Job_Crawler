package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Task struct {
	TaskID         string                 `gorm:"primaryKey" json:"task_id"`          // Task identifier
	ContainerID    string                 `json:"container_id"`                       // Docker Container ID
	DateTime       string                 `json:"date_time"`                          // Task creation time
	Args           map[string]interface{} `gorm:"type:json" json:"args"`              // Task arguments as JSON
	Status         string                 `json:"status"`                             // Task status
	SuccessJobIDs  []string               `gorm:"type:text[]" json:"success_job_ids"` // IDs of successful jobs
	FailedJobIDs   []string               `gorm:"type:text[]" json:"failed_job_ids"`  // IDs of failed jobs
	CompletionRate float64                `json:"completion_rate"`                    // Completion rate
	IsRetryTask    bool                   `json:"is_retry"`                           // Task is retry
	ParentTaskID   string                 `json:"parent_task_id"`                     // Only if task is retried
}

func (t *Task) BeforeSave(tx *gorm.DB) (err error) {
	argsJSON, err := json.Marshal(t.Args)
	if err != nil {
		return err
	}
	tx.Statement.SetColumn("Args", string(argsJSON))
	return nil
}

func (t *Task) AfterFind(tx *gorm.DB) (err error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(t.Args), &args); err != nil {
		return err
	}
	t.Args = args
	return nil
}
