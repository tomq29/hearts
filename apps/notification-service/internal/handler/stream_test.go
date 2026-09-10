package handler_test

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kisssonik/hearts/apps/notification-service/internal/auth"
	"github.com/kisssonik/hearts/apps/notification-service/internal/handler"
	"github.com/kisssonik/hearts/apps/notification-service/internal/hub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	testSecret = "test-secret"
	testOrigin = "http://localhost:5173"
)

func signToken(t *testing.T, userID string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    jwt.NewNumericDate(time.Now().Add(time.Minute)),
		"iss":    "hearts-api",
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	require.NoError(t, err)
	return signed
}

func newServer(t *testing.T) (*httptest.Server, *hub.Hub) {
	t.Helper()
	h := hub.NewHub()
	v, err := auth.NewVerifier(testSecret)
	require.NoError(t, err)

	srv := httptest.NewServer(handler.NewStreamHandler(h, v, zap.NewNop(), []string{testOrigin}))
	t.Cleanup(srv.Close)
	return srv, h
}

// openStream connects and returns a reader over the event stream. The caller
// cancels ctx to close it.
func openStream(t *testing.T, ctx context.Context, url, token string) (*http.Response, *bufio.Reader) {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp, bufio.NewReader(resp.Body)
}

// nextData reads until the next "data:" line, skipping keepalive comments.
func nextData(t *testing.T, r *bufio.Reader) string {
	t.Helper()
	result := make(chan string, 1)
	go func() {
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				result <- ""
				return
			}
			if strings.HasPrefix(line, "data: ") {
				result <- strings.TrimSpace(strings.TrimPrefix(line, "data: "))
				return
			}
		}
	}()
	select {
	case data := <-result:
		return data
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for an event")
		return ""
	}
}

func TestStream_RejectsMissingToken(t *testing.T) {
	srv, _ := newServer(t)

	resp, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestStream_RejectsInvalidToken(t *testing.T) {
	srv, _ := newServer(t)

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// The vulnerability this fixes: the endpoint used to trust ?userId= outright.
func TestStream_QueryParameterCannotImpersonate(t *testing.T) {
	srv, h := newServer(t)

	resp, err := http.Get(srv.URL + "?userId=victim")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "a userId alone must grant nothing")

	// With a real token for "attacker" plus ?userId=victim, the stream must
	// belong to the attacker's own account, not the victim's.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, reader := openStream(t, ctx, srv.URL+"?userId=victim", signToken(t, "attacker"))
	defer stream.Body.Close()
	require.Equal(t, http.StatusOK, stream.StatusCode)

	h.Send("victim", []byte(`{"secret":"for the victim"}`))
	h.Send("attacker", []byte(`{"message":"for the attacker"}`))

	assert.Equal(t, `{"message":"for the attacker"}`, nextData(t, reader),
		"the stream must be keyed by the token's user, not the query parameter")
}

func TestStream_DeliversPublishedEvents(t *testing.T) {
	srv, h := newServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, reader := openStream(t, ctx, srv.URL, signToken(t, "user-1"))
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	h.Send("user-1", []byte(`{"type":"match","message":"You have a new match!"}`))

	assert.Equal(t, `{"type":"match","message":"You have a new match!"}`, nextData(t, reader))
}

// A browser fetch carrying Authorization is preflighted, and the preflight
// never carries credentials, so it must succeed without a token.
func TestStream_AnswersCORSPreflight(t *testing.T) {
	srv, _ := newServer(t)

	req, _ := http.NewRequest(http.MethodOptions, srv.URL, nil)
	req.Header.Set("Origin", testOrigin)
	req.Header.Set("Access-Control-Request-Headers", "authorization")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, testOrigin, resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestStream_UnknownOriginGetsNoCORSHeaders(t *testing.T) {
	srv, _ := newServer(t)

	req, _ := http.NewRequest(http.MethodOptions, srv.URL, nil)
	req.Header.Set("Origin", "http://evil.example.com")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
}
