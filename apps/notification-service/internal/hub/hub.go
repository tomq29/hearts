package hub

import "sync"

// streamBuffer is how many notifications a stream can fall behind by before
// further ones are dropped for it.
const streamBuffer = 16

// Hub routes notifications to the live streams of each user. A user can hold
// several streams at once — one per open tab or device — so each user maps to a
// set of channels rather than a single connection.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[chan []byte]struct{}),
	}
}

// Subscribe registers a new stream for the user. It returns the channel the
// stream receives on, and an unsubscribe func the caller must run when the
// stream ends.
func (h *Hub) Subscribe(userID string) (<-chan []byte, func()) {
	ch := make(chan []byte, streamBuffer)

	h.mu.Lock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[chan []byte]struct{})
	}
	h.clients[userID][ch] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		delete(h.clients[userID], ch)
		if len(h.clients[userID]) == 0 {
			delete(h.clients, userID)
		}
		// The channel is deliberately not closed: the reader exits on its own
		// request context, and not closing rules out a send on a closed channel.
	}

	return ch, unsubscribe
}

// Send delivers a message to every stream the user has open. It never blocks:
// a stream whose buffer is full misses this message rather than stalling the
// Kafka consumer that calls Send.
func (h *Hub) Send(userID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.clients[userID] {
		select {
		case ch <- message:
		default:
		}
	}
}
