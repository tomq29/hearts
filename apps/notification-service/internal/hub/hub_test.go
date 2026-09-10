package hub_test

import (
	"testing"
	"time"

	"github.com/kisssonik/hearts/apps/notification-service/internal/hub"
	"github.com/stretchr/testify/assert"
)

func receive(t *testing.T, ch <-chan []byte) []byte {
	t.Helper()
	select {
	case msg := <-ch:
		return msg
	case <-time.After(time.Second):
		t.Fatal("expected a message, got none")
		return nil
	}
}

// Previously the hub stored one connection per user, so opening a second tab
// silently replaced — and leaked — the first.
func TestSend_ReachesEveryStreamOfTheUser(t *testing.T) {
	h := hub.NewHub()

	tab1, unsub1 := h.Subscribe("user-1")
	defer unsub1()
	tab2, unsub2 := h.Subscribe("user-1")
	defer unsub2()

	h.Send("user-1", []byte("new match"))

	assert.Equal(t, "new match", string(receive(t, tab1)))
	assert.Equal(t, "new match", string(receive(t, tab2)))
}

func TestSend_OnlyReachesTheAddressedUser(t *testing.T) {
	h := hub.NewHub()

	mine, unsubMine := h.Subscribe("user-1")
	defer unsubMine()
	theirs, unsubTheirs := h.Subscribe("user-2")
	defer unsubTheirs()

	h.Send("user-1", []byte("for user-1"))

	assert.Equal(t, "for user-1", string(receive(t, mine)))
	select {
	case msg := <-theirs:
		t.Fatalf("user-2 received a message meant for user-1: %s", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

// One stalled client must not block the Kafka consumer loop that calls Send.
func TestSend_DoesNotBlockOnAFullStream(t *testing.T) {
	h := hub.NewHub()

	_, unsub := h.Subscribe("user-1") // never read from
	defer unsub()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			h.Send("user-1", []byte("flood"))
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Send blocked on a stream nobody was reading")
	}
}

func TestUnsubscribe_StopsDelivery(t *testing.T) {
	h := hub.NewHub()

	ch, unsub := h.Subscribe("user-1")
	unsub()

	h.Send("user-1", []byte("after unsubscribe"))

	select {
	case msg := <-ch:
		t.Fatalf("received after unsubscribe: %s", msg)
	case <-time.After(50 * time.Millisecond):
	}
}
