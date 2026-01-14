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

# Get User A ID
USER_A_ID=$(curl -s -H "Authorization: Bearer $TOKEN_A" $BASE_URL/users/me | jq -r '.id')
echo "User A ID: $USER_A_ID"

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

# Get User B ID
USER_B_ID=$(curl -s -H "Authorization: Bearer $TOKEN_B" $BASE_URL/users/me | jq -r '.id')
echo "User B ID: $USER_B_ID"

# Create Profile B
curl -s -X POST $BASE_URL/profiles \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d '{"firstName": "Bob", "bio": "B", "birthDate": "1990-01-01T00:00:00Z"}' > /dev/null

# 3. Start WebSocket Client for User A
echo "Starting WebSocket listener for User A..."
go run cmd/test_ws/main.go -token "$TOKEN_A" > ws_output.txt 2>&1 &
WS_PID=$!

sleep 2

# 4. User B likes User A
echo "User B likes User A..."
curl -s -X POST $BASE_URL/likes \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d "{\"targetId\": \"$USER_A_ID\", \"isLike\": true}" > /dev/null

# 5. User A likes User B (Match!)
echo "User A likes User B..."
curl -s -X POST $BASE_URL/likes \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d "{\"targetId\": \"$USER_B_ID\", \"isLike\": true}" > /dev/null

echo "Waiting for notification..."
sleep 5

# Kill WebSocket client
kill $WS_PID

# Check output
echo "WebSocket Output:"
cat ws_output.txt

if grep -q "You have a new match!" ws_output.txt; then
    echo "SUCCESS: Notification received via WebSocket"
else
    echo "FAILURE: Notification NOT received"
fi
