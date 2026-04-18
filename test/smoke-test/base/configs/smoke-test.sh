#!/bin/sh
set -euo pipefail

# BASE_URL="http://192.168.254.220"
WEB_BASE_URL="${WEB_BASE_URL:?WEB_BASE_URL environment variable not set}"
API_BASE_URL="${API_BASE_URL:?API_BASE_URL environment variable not set}"

# Helper function for curl with status check
function curl_check() {
  local METHOD=$1
  local URL=$2
  local DATA=${3:-""}
  local AUTH=${4:-""}

  if [ -n "$DATA" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X "$METHOD" "$URL" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $AUTH" \
      -d "$DATA")
  else
    RESPONSE=$(curl -s -w "\n%{http_code}" -X "$METHOD" "$URL" \
      -H "Authorization: Bearer $AUTH")
  fi

  HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
  BODY=$(echo "$RESPONSE" | sed '$d')

  if [ "$HTTP_CODE" -lt 200 ] || [ "$HTTP_CODE" -ge 300 ]; then
    echo "Request to $URL failed with status $HTTP_CODE"
    echo "Response body: $BODY"
    exit 1
  fi

  echo "$BODY"
}

# API Health
printf "\n======** API Health **======\n"
curl_check GET "$API_BASE_URL/api/health"
echo

# Frontend Health (/app)
printf "\n======** Frontend /app Health **======\n"
STATUS=$(curl -s -L -o /dev/null -w "%{http_code}" "$WEB_BASE_URL/app")
if [ "$STATUS" -ne 200 ]; then
  echo "Frontend /app failed with status $STATUS"
  exit 1
fi
echo "status: $STATUS"

# Frontend Health (/status)
printf "\n======** Frontend /status Health **======\n"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$WEB_BASE_URL/status")
if [ "$STATUS" -ne 200 ]; then
  echo "Frontend /status failed with status $STATUS"
  exit 1
fi
echo "status: $STATUS"

# Create dummy accounts
create_user_if_not_exists() {
  local USERNAME=$1
  local PASSWORD=$2

  RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/api/register" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

  HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
  BODY=$(echo "$RESPONSE" | sed '$d')

  if [ "$HTTP_CODE" -ge 200 ] && [ "$HTTP_CODE" -lt 300 ]; then
    echo "User '$USERNAME' created"
  elif echo "$BODY" | grep -qi "exist"; then
    echo "User '$USERNAME' already exists, skipping"
  else
    echo "Failed to create user '$USERNAME' (status $HTTP_CODE)"
    echo "Response: $BODY"
    exit 1
  fi
}

printf "\n======** Create User 'admin' **======\n"
# curl_check POST "$API_BASE_URL/api/register" '{"username":"admin","password":"admin"}'
create_user_if_not_exists "admin" "admin"
echo

printf "\n======** Create User 'jonathan' **======\n"
curl_check POST "$API_BASE_URL/api/register" '{"username":"jonathan","password":"123"}'
echo

# Get token for 'admin'
printf "\n======** Get Admin Token **======\n"
TOKEN=$(curl_check POST "$API_BASE_URL/api/login" '{"username":"admin","password":"admin"}' | jq -r '.token')
if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
  echo "Failed to obtain JWT token"
  exit 1
fi
echo "Token: $TOKEN"

# List users
printf "\n======** List Users **======\n"
curl_check GET "$API_BASE_URL/api/users" "" "$TOKEN" | jq
echo

# Show version
printf "\n======** API Version **======\n"
curl_check GET "$API_BASE_URL/api/version" "" "$TOKEN" | jq
echo

# Update user password
printf "\n======** Update Password for 'jonathan' **======\n"
curl_check PUT "$API_BASE_URL/api/users/jonathan" '{"password":"456"}' "$TOKEN"
echo

# Delete user
printf "\n======** Delete User 'jonathan' **======\n"
curl_check DELETE "$API_BASE_URL/api/users/jonathan" "" "$TOKEN"
echo
