#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

# 1. Register User A
echo "Registering User A..."
# Use a random suffix to avoid conflicts on re-runs
SUFFIX=$RANDOM
EMAIL_A="userA_$SUFFIX@example.com"
USERNAME_A="userA_$SUFFIX"

TOKEN_A=$(curl -s -X POST $BASE_URL/users/register \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_A\", \"password\": \"password123\", \"username\": \"$USERNAME_A\"}" | jq -r '.token')

# If registration returns user object (no token), we need to login
echo "Login User A..."
TOKEN_A=$(curl -s -X POST $BASE_URL/users/login \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_A\", \"password\": \"password123\"}" | jq -r '.accessToken')

echo "Token A: $TOKEN_A"

# Create Profile A
echo "Creating Profile A..."
curl -s -X POST $BASE_URL/profiles \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{"firstName": "Alice", "bio": "Hi"}' > /dev/null

# 2. Register User B
echo "Registering User B..."
EMAIL_B="userB_$SUFFIX@example.com"
USERNAME_B="userB_$SUFFIX"

curl -s -X POST $BASE_URL/users/register \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_B\", \"password\": \"password123\", \"username\": \"$USERNAME_B\"}" > /dev/null

echo "Login User B..."
TOKEN_B=$(curl -s -X POST $BASE_URL/users/login \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_B\", \"password\": \"password123\"}" | jq -r '.accessToken')

echo "Token B: $TOKEN_B"

# Create Profile B
echo "Creating Profile B..."
curl -s -X POST $BASE_URL/profiles \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d '{"firstName": "Bob", "bio": "Hello"}' > /dev/null

# Get IDs
ID_A=$(curl -s -H "Authorization: Bearer $TOKEN_A" $BASE_URL/users/me | jq -r '.id')
ID_B=$(curl -s -H "Authorization: Bearer $TOKEN_B" $BASE_URL/users/me | jq -r '.id')

echo "ID A: $ID_A"
echo "ID B: $ID_B"

# 3. User A likes User B
echo "User A likes User B..."
curl -s -X POST $BASE_URL/likes \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d "{\"targetId\": \"$ID_B\", \"isLike\": true}"

# 4. User B likes User A
echo "User B likes User A..."
curl -s -X POST $BASE_URL/likes \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d "{\"targetId\": \"$ID_A\", \"isLike\": true}"

echo "Waiting for Kafka worker to process..."
sleep 5

# 5. Check Matches for User A
echo "Checking matches for User A..."
MATCHES=$(curl -s -H "Authorization: Bearer $TOKEN_A" $BASE_URL/matches)
echo $MATCHES

if echo "$MATCHES" | grep -q "$ID_B"; then
    echo "SUCCESS: Match found!"
else
    echo "FAILURE: Match not found."
fi
