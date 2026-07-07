package main

import "time"

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusClosed     = "closed"
)

type Ticket struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var allowedTransitions = map[string]string{
	StatusOpen:       StatusInProgress,
	StatusInProgress: StatusClosed,
}

func isValidStatus(s string) bool {
	return s == StatusOpen || s == StatusInProgress || s == StatusClosed
}

func canTransition(from, to string) bool {
	if from == StatusClosed {
		return false
	}
	return allowedTransitions[from] == to
}
