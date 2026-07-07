package main

import (
	"sync"
)

type Store struct {
	mu           sync.RWMutex
	usersByEmail map[string]*User
	usersByID    map[string]*User
	tickets      map[string]*Ticket
	nextUserID   int
	nextTicketID int
}

func NewStore() *Store {
	return &Store{
		usersByEmail: make(map[string]*User),
		usersByID:    make(map[string]*User),
		tickets:      make(map[string]*Ticket),
	}
}

func (s *Store) CreateUser(email, passwordHash string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByEmail[email]; exists {
		return nil, ErrEmailTaken
	}

	s.nextUserID++
	u := &User{
		ID:           genID("usr", s.nextUserID),
		Email:        email,
		PasswordHash: passwordHash,
	}
	s.usersByEmail[email] = u
	s.usersByID[u.ID] = u
	return u, nil
}

func (s *Store) GetUserByEmail(email string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.usersByEmail[email]
	return u, ok
}

func (s *Store) GetUserByID(id string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.usersByID[id]
	return u, ok
}

func (s *Store) CreateTicket(userID, title, description string) *Ticket {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextTicketID++
	now := nowUTC()
	t := &Ticket{
		ID:          genID("tkt", s.nextTicketID),
		UserID:      userID,
		Title:       title,
		Description: description,
		Status:      StatusOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.tickets[t.ID] = t
	return t
}

func (s *Store) ListTicketsByUser(userID string) []*Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Ticket
	for _, t := range s.tickets {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	return result
}

func (s *Store) GetTicket(id string) (*Ticket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickets[id]
	return t, ok
}

func (s *Store) UpdateTicketStatus(id, newStatus string) (*Ticket, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tickets[id]
	if !ok {
		return nil, false
	}
	t.Status = newStatus
	t.UpdatedAt = nowUTC()
	return t, true
}
