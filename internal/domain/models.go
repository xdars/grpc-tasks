package domain

import "time"

type TaskStatus int32
type PaymentStatus int32

const (
	PaymentStatusUnspecified PaymentStatus = 0
	PaymentStatusPending     PaymentStatus = 1
	PaymentStatusProcessing  PaymentStatus = 2
	PaymentStatusSuccess     PaymentStatus = 3
	PaymentStatusFailed      PaymentStatus = 4
)

type PaymentType int32

const (
	PaymentTypeUnspecified PaymentType = 0
	PaymentTypePremium     PaymentType = 1
	PaymentTypeBonus       PaymentType = 2
)

const (
	TaskStatusUnspecified TaskStatus = 0
	TaskStatusTodo        TaskStatus = 1
	TaskStatusInProgress  TaskStatus = 2
	TaskStatusDone        TaskStatus = 3
)

type Task struct {
	ID          string
	UserID      string
	Title       string
	Description string
	Status      TaskStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type User struct {
	ID           string
	Username     string
	PasswordHash string
}

type Payment struct {
	ID        string
	UserID    string
	Type      PaymentType
	Status    PaymentStatus
	Amount    int64
	Currency  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
