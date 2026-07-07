package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"ticket-system/internal/middleware"
	"ticket-system/internal/models"
	"ticket-system/internal/store"
)

type TicketHandler struct {
	store *store.Store
}

func NewTicketHandler(s *store.Store) *TicketHandler {
	return &TicketHandler{store: s}
}

func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	ticket := h.store.CreateTicket(user.UserID, req.Title, req.Description)
	writeJSON(w, http.StatusCreated, ticket)
}

func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tickets := h.store.ListTicketsByUser(user.UserID)
	if tickets == nil {
		tickets = []*models.Ticket{}
	}

	writeJSON(w, http.StatusOK, tickets)
}

func (h *TicketHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/tickets/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	ticket, err := h.store.GetTicket(id)
	if err != nil {
		if errors.Is(err, store.ErrTicketNotFound) {
			writeError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get ticket")
		return
	}

	if ticket.UserID != user.UserID {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

func (h *TicketHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/tickets/")
	path = strings.TrimSuffix(path, "/status")
	if path == "" || strings.Contains(path, "/") {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	ticket, err := h.store.GetTicket(path)
	if err != nil {
		if errors.Is(err, store.ErrTicketNotFound) {
			writeError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get ticket")
		return
	}

	if ticket.UserID != user.UserID {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validateStatusTransition(ticket.Status, req.Status); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.store.UpdateTicketStatus(path, req.Status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update ticket")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func validateStatusTransition(current, next models.TicketStatus) error {
	if current == next {
		return errors.New("ticket is already in the requested status")
	}

	switch current {
	case models.StatusOpen:
		if next != models.StatusInProgress {
			return errors.New("invalid status transition: open can only move to in_progress")
		}
	case models.StatusInProgress:
		if next != models.StatusClosed {
			return errors.New("invalid status transition: in_progress can only move to closed")
		}
	case models.StatusClosed:
		return errors.New("closed tickets cannot be reopened")
	default:
		return errors.New("invalid current status")
	}

	if next != models.StatusOpen && next != models.StatusInProgress && next != models.StatusClosed {
		return errors.New("invalid status value")
	}

	return nil
}
