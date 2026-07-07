package store

import (
	"errors"
	"sync"
	"time"

	"ticket-system/internal/models"

	"github.com/google/uuid"
)

var (
	ErrUserExists    = errors.New("user already exists")
	ErrUserNotFound  = errors.New("user not found")
	ErrTicketNotFound = errors.New("ticket not found")
)

type Store struct {
	mu      sync.RWMutex
	users   map[string]*models.User
	usersByEmail map[string]string
	tickets map[string]*models.Ticket
}

func New() *Store {
	return &Store{
		users:        make(map[string]*models.User),
		usersByEmail: make(map[string]string),
		tickets:      make(map[string]*models.Ticket),
	}
}

func (s *Store) CreateUser(email, passwordHash string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByEmail[email]; exists {
		return nil, ErrUserExists
	}

	now := time.Now().UTC()
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
	}

	s.users[user.ID] = user
	s.usersByEmail[email] = user.ID
	return user, nil
}

func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, ok := s.usersByEmail[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return s.users[userID], nil
}

func (s *Store) CreateTicket(userID, title, description string) *models.Ticket {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	ticket := &models.Ticket{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       title,
		Description: description,
		Status:      models.StatusOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.tickets[ticket.ID] = ticket
	return ticket
}

func (s *Store) ListTicketsByUser(userID string) []*models.Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.Ticket, 0)
	for _, ticket := range s.tickets {
		if ticket.UserID == userID {
			result = append(result, ticket)
		}
	}
	return result
}

func (s *Store) GetTicket(id string) (*models.Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ticket, ok := s.tickets[id]
	if !ok {
		return nil, ErrTicketNotFound
	}
	return ticket, nil
}

func (s *Store) UpdateTicketStatus(id string, status models.TicketStatus) (*models.Ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ticket, ok := s.tickets[id]
	if !ok {
		return nil, ErrTicketNotFound
	}

	ticket.Status = status
	ticket.UpdatedAt = time.Now().UTC()
	return ticket, nil
}
