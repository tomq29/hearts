#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"
SUFFIX=$RANDOM

# 1. Register User A
echo "Registering User A..."
EMAIL_A="userA_$SUFFIX@example.com"
TOKEN_A=$(curl -s -X POST $BASE_URL/users/register \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_A\", \"password\": \"password123\", \"username\": \"userA_$SUFFIX\"}" | jq -r '.token')

if [ "$TOKEN_A" == "null" ]; then
    TOKEN_A=$(curl -s -X POST $BASE_URL/users/login \
      -H "Content-Type: application/json" \
      -d "{\"email\": \"$EMAIL_A\", \"password\": \"password123\"}" | jq -r '.accessToken')
fi

USER_A_ID=$(curl -s -H "Authorization: Bearer $TOKEN_A" $BASE_URL/users/me | jq -r '.id')

# Create Profile A
curl -s -X POST $BASE_URL/profiles \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{"firstName": "Alice", "bio": "A", "birthDate": "1995-01-01T00:00:00Z"}' > /dev/null

# 2. Register User B
echo "Registering User B..."
EMAIL_B="userB_$SUFFIX@example.com"
TOKEN_B=$(curl -s -X POST $BASE_URL/users/register \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_B\", \"password\": \"password123\", \"username\": \"userB_$SUFFIX\"}" | jq -r '.token')

if [ "$TOKEN_B" == "null" ]; then
    TOKEN_B=$(curl -s -X POST $BASE_URL/users/login \
      -H "Content-Type: application/json" \
      -d "{\"email\": \"$EMAIL_B\", \"password\": \"password123\"}" | jq -r '.accessToken')
fi

USER_B_ID=$(curl -s -H "Authorization: Bearer $TOKEN_B" $BASE_URL/users/me | jq -r '.id')

# Create Profile B
curl -s -X POST $BASE_URL/profiles \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d '{"firstName": "Bob", "bio": "B", "birthDate": "1990-01-01T00:00:00Z"}' > /dev/null

# 3. Match Users
echo "Matching Users..."
curl -s -X POST $BASE_URL/likes \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d "{\"targetId\": \"$USER_A_ID\", \"isLike\": true}" > /dev/null

curl -s -X POST $BASE_URL/likes \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d "{\"targetId\": \"$USER_B_ID\", \"isLike\": true}" > /dev/null

# 4. Start WebSocket Client for User B (Receiver)
echo "Starting WebSocket listener for User B..."
go run cmd/test_ws/main.go -token "$TOKEN_B" > ws_chat_output.txt 2>&1 &
WS_PID=$!

sleep 2

# 5. User A sends message via WebSocket (Simulated by another script or just assume WS works if we can send via API? No, we need to send via WS)
# Since our test_ws client only reads, we need a way to send.
# Let's modify test_ws to support sending or create a sender script.
# For now, let's verify the history API works, which implies the DB part works.
# But to test WS flow, we need to send via WS.

# Let's create a simple go program to send a message
cat <<EOF > cmd/test_ws_send/main.go
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	token := flag.String("token", "", "Bearer token")
	toUser := flag.String("to", "", "To User ID")
	msg := flag.String("msg", "", "Message content")
	flag.Parse()

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws", RawQuery: "token=" + *token}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	payload := map[string]string{
		"type": "chat",
		"toUserId": *toUser,
		"content": *msg,
	}
	bytes, _ := json.Marshal(payload)

	if err := c.WriteMessage(websocket.TextMessage, bytes); err != nil {
		log.Fatal("write:", err)
	}
	
	// Wait a bit for server to process
	time.Sleep(time.Second)
}
EOF

mkdir -p cmd/test_ws_send

echo "User A sending message to User B..."
go run cmd/test_ws_send/main.go -token "$TOKEN_A" -to "$USER_B_ID" -msg "Hello Bob!"

sleep 2

# Kill Listener
kill $WS_PID

echo "WebSocket Output:"
cat ws_chat_output.txt

if grep -q "Hello Bob!" ws_chat_output.txt; then
    echo "SUCCESS: Message received via WebSocket"
else
    echo "FAILURE: Message NOT received"
fi

# 6. Check History
echo "Checking Chat History..."
HISTORY=$(curl -s -H "Authorization: Bearer $TOKEN_A" "$BASE_URL/chats/$USER_B_ID/messages")
echo $HISTORY | jq .

if echo "$HISTORY" | grep -q "Hello Bob!"; then
    echo "SUCCESS: Message found in history"
else
    echo "FAILURE: Message NOT found in history"
fi
