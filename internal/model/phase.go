package model

import (
	"time"
)

type TaskStatus string

const (
	StatusPending         TaskStatus = "PENDING"
	StatusWorking         TaskStatus = "WORKING"
	StatusTesting         TaskStatus = "TESTING"
	StatusDone            TaskStatus = "DONE"
	StatusBlocked         TaskStatus = "BLOCKED"
	StatusFailed          TaskStatus = "FAILED"
	StatusWaitingApproval TaskStatus = "WAITING_APPROVAL"
	StatusPlanned         TaskStatus = "PLANNED"
	StatusSuperseded      TaskStatus = "SUPERSEDED"
)

type PhaseStatus string

const (
	PhasePending    PhaseStatus = "PENDING"
	PhaseWorking    PhaseStatus = "WORKING"
	PhaseCompleted  PhaseStatus = "COMPLETED"
)

type Priority string

const (
	PriorityLow      Priority = "LOW"
	PriorityMedium   Priority = "MEDIUM"
	PriorityHigh     Priority = "HIGH"
	PriorityCritical Priority = "CRITICAL"
)

type DevelopmentPhase struct {
	ID            uint                  `gorm:"primaryKey" json:"id"`
	PhaseNumber   int                   `gorm:"uniqueIndex;not null" json:"phase_number"`
	Name          string                `gorm:"size:128;not null" json:"name"`
	Description   string                `gorm:"type:text" json:"description"`
	Status        PhaseStatus           `gorm:"size:32;default:'PENDING'" json:"status"`
	Progress      float64               `gorm:"default:0" json:"progress"`
	OrderIndex    int                   `gorm:"default:0" json:"order_index"`
	TargetDate    *time.Time            `json:"target_date"`
	CompletedDate *time.Time            `json:"completed_date"`
	Subphases     []DevelopmentSubphase `gorm:"foreignKey:PhaseID" json:"subphases,omitempty"`
	Tasks         []DevelopmentTask     `gorm:"foreignKey:PhaseID" json:"tasks,omitempty"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

type DevelopmentSubphase struct {
	ID                 uint              `gorm:"primaryKey" json:"id"`
	PhaseID            uint              `gorm:"index;not null" json:"phase_id"`
	Name               string            `gorm:"size:128;not null" json:"name"`
	Description        string            `gorm:"type:text" json:"description"`
	Status             string            `gorm:"size:32;default:'PENDING'" json:"status"`
	AcceptanceCriteria string            `gorm:"size:64" json:"acceptance_criteria"`
	OrderIndex         int               `gorm:"default:0" json:"order_index"`
	Progress           float64           `gorm:"default:0" json:"progress"`
	Tasks              []DevelopmentTask `gorm:"foreignKey:SubphaseID" json:"tasks,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

type DevelopmentTask struct {
	ID             uint                    `gorm:"primaryKey" json:"id"`
	PhaseID        uint                    `gorm:"index;not null" json:"phase_id"`
	SubphaseID     *uint                   `gorm:"index" json:"subphase_id"`
	TaskName       string                  `gorm:"size:128;not null" json:"task_name"`
	Description    string                  `gorm:"type:text" json:"description"`
	Status         TaskStatus              `gorm:"size:32;default:'PENDING'" json:"status"`
	Progress       float64                 `gorm:"default:0" json:"progress"`
	Priority       Priority                `gorm:"size:16;default:'MEDIUM'" json:"priority"`
	Owner          string                  `gorm:"size:64;default:'Engineer'" json:"owner"`
	StartDate      *time.Time              `json:"start_date"`
	DueDate        *time.Time              `json:"due_date"`
	CompletionDate *time.Time              `json:"completion_date"`
	Notes          string                  `gorm:"type:text" json:"notes"`
	TestResult     string                  `gorm:"type:text" json:"test_result"`
	OrderIndex     int                     `gorm:"default:0" json:"order_index"`
	Logs           []DevelopmentTaskLog    `gorm:"foreignKey:TaskID" json:"logs,omitempty"`
	Evidences      []DevelopmentEvidence   `gorm:"foreignKey:TaskID" json:"evidences,omitempty"`
	Dependencies   []DevelopmentDependency `gorm:"foreignKey:TaskID" json:"dependencies,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

type DevelopmentTaskLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    uint      `gorm:"index;not null" json:"task_id"`
	Timestamp time.Time `gorm:"not null" json:"timestamp"`
	User      string    `gorm:"size:64" json:"user"`
	Action    string    `gorm:"size:128;not null" json:"action"`
	Result    string    `gorm:"size:255" json:"result"`
	Log       string    `gorm:"type:text" json:"log"`
	CreatedAt time.Time `json:"created_at"`
}

type DevelopmentDependency struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	TaskID          uint             `gorm:"index;not null" json:"task_id"`
	DependsOnTaskID uint             `gorm:"index;not null" json:"depends_on_task_id"`
	DependsOnTask   *DevelopmentTask `gorm:"foreignKey:DependsOnTaskID" json:"depends_on_task,omitempty"`
}

type DevelopmentEvidence struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TaskID      uint      `gorm:"index;not null" json:"task_id"`
	Title       string    `gorm:"size:128;not null" json:"title"`
	FilePath    string    `gorm:"size:255" json:"file_path"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
