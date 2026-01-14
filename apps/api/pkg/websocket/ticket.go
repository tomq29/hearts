package websocket

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type TicketStore struct {
	tickets map[string]string // ticket -> userID
	mu      sync.RWMutex
}

func NewTicketStore() *TicketStore {
	return &TicketStore{
		tickets: make(map[string]string),
	}
}

func (ts *TicketStore) Create(userID string) string {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ticket := uuid.New().String()
	ts.tickets[ticket] = userID

	// Auto-expire after 30 seconds
	time.AfterFunc(30*time.Second, func() {
		ts.mu.Lock()
		delete(ts.tickets, ticket)
		ts.mu.Unlock()
	})

	return ticket
}

func (ts *TicketStore) Validate(ticket string) (string, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	userID, ok := ts.tickets[ticket]
	if ok {
		delete(ts.tickets, ticket) // One-time use
	}
	return userID, ok
}
