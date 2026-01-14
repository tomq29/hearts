#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"
SUFFIX=$RANDOM

# 1. Register User A (NYC)
echo "Registering User A (NYC)..."
EMAIL_A="userA_$SUFFIX@example.com"
TOKEN_A=$(curl -s -X POST $BASE_URL/users/register \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_A\", \"password\": \"password123\", \"username\": \"userA_$SUFFIX\"}" | jq -r '.token')

if [ "$TOKEN_A" == "null" ]; then
    TOKEN_A=$(curl -s -X POST $BASE_URL/users/login \
      -H "Content-Type: application/json" \
      -d "{\"email\": \"$EMAIL_A\", \"password\": \"password123\"}" | jq -r '.accessToken')
fi

# Create Profile A (NYC, Born 1995)
curl -s -X POST $BASE_URL/profiles \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{"firstName": "Alice", "bio": "NYC", "latitude": 40.7128, "longitude": -74.0060, "birthDate": "1995-01-01T00:00:00Z"}' > /dev/null

# 2. Register User B (Nearby - ~5km away, Born 1990)
echo "Registering User B (Nearby)..."
EMAIL_B="userB_$SUFFIX@example.com"
curl -s -X POST $BASE_URL/users/register \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_B\", \"password\": \"password123\", \"username\": \"userB_$SUFFIX\"}" > /dev/null

TOKEN_B=$(curl -s -X POST $BASE_URL/users/login \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_B\", \"password\": \"password123\"}" | jq -r '.accessToken')

# Create Profile B
curl -s -X POST $BASE_URL/profiles \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d '{"firstName": "Bob", "bio": "Nearby", "latitude": 40.7306, "longitude": -73.9352, "birthDate": "1990-01-01T00:00:00Z"}' > /dev/null

# 3. Register User C (Far - LA, Born 2000)
echo "Registering User C (Far)..."
EMAIL_C="userC_$SUFFIX@example.com"
curl -s -X POST $BASE_URL/users/register \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_C\", \"password\": \"password123\", \"username\": \"userC_$SUFFIX\"}" > /dev/null

TOKEN_C=$(curl -s -X POST $BASE_URL/users/login \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_C\", \"password\": \"password123\"}" | jq -r '.accessToken')

# Create Profile C
curl -s -X POST $BASE_URL/profiles \
  -H "Authorization: Bearer $TOKEN_C" \
  -H "Content-Type: application/json" \
  -d '{"firstName": "Charlie", "bio": "Far", "latitude": 34.0522, "longitude": -118.2437, "birthDate": "2000-01-01T00:00:00Z"}' > /dev/null

echo "--- Test 1: Search Radius 10km (Should find Bob) ---"
SEARCH_1=$(curl -s -H "Authorization: Bearer $TOKEN_A" "$BASE_URL/profiles/search?radius=10")
echo $SEARCH_1 | jq .

if echo "$SEARCH_1" | grep -q "Bob"; then
    echo "SUCCESS: Found Bob"
else
    echo "FAILURE: Did not find Bob"
fi

if echo "$SEARCH_1" | grep -q "Charlie"; then
    echo "FAILURE: Found Charlie (should be too far)"
else
    echo "SUCCESS: Did not find Charlie"
fi

echo "--- Test 2: Search Radius 1km (Should find no one) ---"
SEARCH_2=$(curl -s -H "Authorization: Bearer $TOKEN_A" "$BASE_URL/profiles/search?radius=1")
echo $SEARCH_2 | jq .

if echo "$SEARCH_2" | grep -q "Bob"; then
    echo "FAILURE: Found Bob (should be too far)"
else
    echo "SUCCESS: Did not find Bob"
fi

echo "--- Test 3: Search Age (Min 30) (Should find Bob (35), not Charlie (25)) ---"
# Bob is ~35 (1990), Charlie is ~25 (2000)
SEARCH_3=$(curl -s -H "Authorization: Bearer $TOKEN_A" "$BASE_URL/profiles/search?minAge=30")
echo $SEARCH_3 | jq .

if echo "$SEARCH_3" | grep -q "Bob"; then
    echo "SUCCESS: Found Bob"
else
    echo "FAILURE: Did not find Bob"
fi

if echo "$SEARCH_3" | grep -q "Charlie"; then
    echo "FAILURE: Found Charlie (should be too young)"
else
    echo "SUCCESS: Did not find Charlie"
fi
