package database

import (
	"time"
)

type Task struct {
	UUID                              string    `json:"uuid"`
	Body                              string    `json:"body"`
	TimeCreated                       time.Time `json:"timecreated"`
	TimeCompleted                     time.Time `json:"timecompleted"`
	DurationExecutionEstimatedSeconds int       `json:"duration_execution_estimated_seconds"`
	DurationExecutionRealSeconds      int       `json:"duration_execution_real_seconds"`
	TimeHardDeadline                  time.Time `json:"time_hard_dead_line"`
	Order                             int       `json:"order"`
	Child                             []Task    `json:"-"`
}

type Gamification struct {
	TotalPoints      int       `json:"total_points"`
	CurrentStreak    int       `json:"current_streak"`
	LongestStreak    int       `json:"longest_streak"`
	LastCompletionDate *time.Time `json:"last_completion_date"`
	Level            int       `json:"level"`
	CompletedTasks   int       `json:"completed_tasks"`
	FirstTaskDate    *time.Time `json:"first_task_date"`
	Achievements     []string  `json:"achievements"`
}

type Database interface {
	Connect() error
	Disconnect() error

	GetTasks() ([]Task, error)
	GetTaskByUUID(uuid string) (*Task, error)
	AddTask(task *Task) error
	UpdateTask(task *Task) error
	RemoveTask(uuid string) error

	GetCompletedTasks() ([]Task, error)
	AddCompletedTask(task *Task) error

	GetGamification() (*Gamification, error)
	UpdateGamification(gamification *Gamification) error

	DBUpgrade() string
}

