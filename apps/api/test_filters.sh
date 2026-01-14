#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"
SUFFIX=$RANDOM

# 1. Register User A (Female, 160cm)
echo "Registering User A (Female, 160cm)..."
EMAIL_A="userA_$SUFFIX@example.com"
TOKEN_A=$(curl -s -X POST $BASE_URL/users/register \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_A\", \"password\": \"password123\", \"username\": \"userA_$SUFFIX\"}" | jq -r '.token')

if [ "$TOKEN_A" == "null" ]; then
    TOKEN_A=$(curl -s -X POST $BASE_URL/users/login \
      -H "Content-Type: application/json" \
      -d "{\"email\": \"$EMAIL_A\", \"password\": \"password123\"}" | jq -r '.accessToken')
fi

# Create Profile A
curl -s -X POST $BASE_URL/profiles \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{"firstName": "Alice", "bio": "F 160", "gender": "female", "height": 160, "birthDate": "1995-01-01T00:00:00Z"}' > /dev/null

# 2. Register User B (Male, 180cm)
echo "Registering User B (Male, 180cm)..."
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
  -d '{"firstName": "Bob", "bio": "M 180", "gender": "male", "height": 180, "birthDate": "1990-01-01T00:00:00Z"}' > /dev/null

# 3. Register User C (Male, 170cm)
echo "Registering User C (Male, 170cm)..."
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
  -d '{"firstName": "Charlie", "bio": "M 170", "gender": "male", "height": 170, "birthDate": "2000-01-01T00:00:00Z"}' > /dev/null

echo "--- Test 1: Search Gender 'male' (Should find Bob and Charlie) ---"
SEARCH_1=$(curl -s -H "Authorization: Bearer $TOKEN_A" "$BASE_URL/profiles/search?gender=male")
echo $SEARCH_1 | jq .

if echo "$SEARCH_1" | grep -q "Bob" && echo "$SEARCH_1" | grep -q "Charlie"; then
    echo "SUCCESS: Found Bob and Charlie"
else
    echo "FAILURE: Did not find Bob and Charlie"
fi

if echo "$SEARCH_1" | grep -q "Alice"; then
    echo "FAILURE: Found Alice (should be filtered out)"
else
    echo "SUCCESS: Did not find Alice"
fi

echo "--- Test 2: Search Height >= 175 (Should find Bob) ---"
SEARCH_2=$(curl -s -H "Authorization: Bearer $TOKEN_A" "$BASE_URL/profiles/search?minHeight=175")
echo $SEARCH_2 | jq .

if echo "$SEARCH_2" | grep -q "Bob"; then
    echo "SUCCESS: Found Bob"
else
    echo "FAILURE: Did not find Bob"
fi

if echo "$SEARCH_2" | grep -q "Charlie"; then
    echo "FAILURE: Found Charlie (should be filtered out)"
else
    echo "SUCCESS: Did not find Charlie"
fi

echo "--- Test 3: Search Height <= 165 (Should find Alice - but searching as Bob) ---"
SEARCH_3=$(curl -s -H "Authorization: Bearer $TOKEN_B" "$BASE_URL/profiles/search?maxHeight=165")
echo $SEARCH_3 | jq .

if echo "$SEARCH_3" | grep -q "Alice"; then
    echo "SUCCESS: Found Alice"
else
    echo "FAILURE: Did not find Alice"
fi

if echo "$SEARCH_3" | grep -q "Charlie"; then
    echo "FAILURE: Found Charlie (should be filtered out)"
else
    echo "SUCCESS: Did not find Charlie"
fi
