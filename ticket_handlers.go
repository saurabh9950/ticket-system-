package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type createTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (app *App) handleCreateTicket(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)

	var req createTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	ticket := app.store.CreateTicket(userID, req.Title, req.Description)
	writeJSON(w, http.StatusCreated, ticket)
}

func (app *App) handleListTickets(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	tickets := app.store.ListTicketsByUser(userID)
	if tickets == nil {
		tickets = []*Ticket{}
	}
	writeJSON(w, http.StatusOK, tickets)
}

func (app *App) handleGetTicket(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	id := r.PathValue("id")

	ticket, ok := app.store.GetTicket(id)
	if !ok {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}
	if ticket.UserID != userID {
		writeError(w, http.StatusForbidden, "you do not have access to this ticket")
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

func (app *App) handleUpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	id := r.PathValue("id")

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if !isValidStatus(req.Status) {
		writeError(w, http.StatusBadRequest, "status must be one of: open, in_progress, closed")
		return
	}

	ticket, ok := app.store.GetTicket(id)
	if !ok {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}
	if ticket.UserID != userID {
		writeError(w, http.StatusForbidden, "you do not have access to this ticket")
		return
	}

	if !canTransition(ticket.Status, req.Status) {
		writeError(w, http.StatusBadRequest, "invalid status transition: "+ticket.Status+" -> "+req.Status)
		return
	}

	updated, _ := app.store.UpdateTicketStatus(id, req.Status)
	writeJSON(w, http.StatusOK, updated)
}
