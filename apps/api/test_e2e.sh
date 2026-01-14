#!/bin/bash

API_URL="http://localhost:8080/api/v1"
EMAIL_1="user1_$(date +%s)@example.com"
EMAIL_2="user2_$(date +%s)@example.com"
PASS="password123"

echo "--- 1. Register User 1 ---"
curl -v -X POST "$API_URL/users/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_1\", \"username\": \"user1_$(date +%s)\", \"password\": \"$PASS\"}"
echo -e "\n"

echo "--- 2. Login User 1 ---"
TOKEN_1=$(curl -s -X POST "$API_URL/users/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_1\", \"password\": \"$PASS\"}" | jq -r '.accessToken')
echo "Token 1: ${TOKEN_1:0:10}..."

echo "--- 3. Create Profile User 1 ---"
curl -s -X POST "$API_URL/profiles" \
  -H "Authorization: Bearer $TOKEN_1" \
  -H "Content-Type: application/json" \
  -d '{"firstName": "Alice", "bio": "Loves crypto", "selfDescribedStrengths": ["Smart"], "selfDescribedFlaws": ["Impulsive"]}'
echo -e "\n"

echo "--- 4. Upload Photo User 1 ---"
# Assuming test_image.jpg exists
curl -s -X POST "$API_URL/profiles/upload" \
  -H "Authorization: Bearer $TOKEN_1" \
  -F "photo=@test_image.jpg"
echo -e "\n"

echo "--- 5. Register User 2 ---"
curl -s -X POST "$API_URL/users/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_2\", \"username\": \"user2_$(date +%s)\", \"password\": \"$PASS\"}"
echo -e "\n"

echo "--- 6. Login User 2 ---"
TOKEN_2=$(curl -s -X POST "$API_URL/users/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL_2\", \"password\": \"$PASS\"}" | jq -r '.accessToken')
echo "Token 2: ${TOKEN_2:0:10}..."

echo "--- 7. Create Profile User 2 ---"
curl -s -X POST "$API_URL/profiles" \
  -H "Authorization: Bearer $TOKEN_2" \
  -H "Content-Type: application/json" \
  -d '{"firstName": "Bob", "bio": "Loves cats", "selfDescribedStrengths": ["Kind"], "selfDescribedFlaws": ["Lazy"]}'
echo -e "\n"

echo "--- 8. Get User 1 ID (Me) ---"
USER_1_ID=$(curl -s -X GET "$API_URL/users/me" -H "Authorization: Bearer $TOKEN_1" | jq -r '.id')
echo "User 1 ID: $USER_1_ID"

echo "--- 9. Get User 2 ID (Me) ---"
USER_2_ID=$(curl -s -X GET "$API_URL/users/me" -H "Authorization: Bearer $TOKEN_2" | jq -r '.id')
echo "User 2 ID: $USER_2_ID"

echo "--- 10. User 1 Likes User 2 ---"
curl -s -X POST "$API_URL/likes" \
  -H "Authorization: Bearer $TOKEN_1" \
  -H "Content-Type: application/json" \
  -d "{\"targetId\": \"$USER_2_ID\", \"isLike\": true}"
echo -e "\n"

echo "--- 11. User 2 Likes User 1 (MATCH!) ---"
curl -s -X POST "$API_URL/likes" \
  -H "Authorization: Bearer $TOKEN_2" \
  -H "Content-Type: application/json" \
  -d "{\"targetId\": \"$USER_1_ID\", \"isLike\": true}"
echo -e "\n"

echo "--- 12. Check Matches for User 1 ---"
curl -s -X GET "$API_URL/matches" \
  -H "Authorization: Bearer $TOKEN_1" | jq .
echo -e "\n"

echo "--- 13. User 1 Reviews User 2 ---"
curl -s -X POST "$API_URL/reviews" \
  -H "Authorization: Bearer $TOKEN_1" \
  -H "Content-Type: application/json" \
  -d "{\"targetId\": \"$USER_2_ID\", \"rating\": 5, \"comment\": \"Great match!\"}"
echo -e "\n"

echo "--- 14. Get Reviews for User 2 ---"
curl -s -X GET "$API_URL/users/$USER_2_ID/reviews" | jq .
echo -e "\n"
