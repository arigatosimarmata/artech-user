#!/bin/bash

# API Testing Script for Artech User Service
# This script tests all available endpoints

BASE_URL="http://localhost:8080"

echo "================================"
echo "Artech User Service API Tests"
echo "================================"
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Health Check
echo -e "${YELLOW}Test 1: Health Check${NC}"
curl -s -X GET $BASE_URL/health | jq .
echo ""

# Test 2: Register User
echo -e "${YELLOW}Test 2: Register User${NC}"
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "demo123456",
    "full_name": "Demo User",
    "phone": "1234567890"
  }')
echo $REGISTER_RESPONSE | jq .
ACCESS_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.data.access_token')
REFRESH_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.data.refresh_token')
echo ""

# Test 3: Get Profile (Protected)
echo -e "${YELLOW}Test 3: Get Profile (Protected)${NC}"
curl -s -X GET $BASE_URL/api/v1/user/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Test 4: Login
echo -e "${YELLOW}Test 4: Login${NC}"
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "demo123456"
  }')
echo $LOGIN_RESPONSE | jq .
NEW_ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.access_token')
NEW_REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.refresh_token')
echo ""

# Test 5: Refresh Token
echo -e "${YELLOW}Test 5: Refresh Token${NC}"
curl -s -X POST $BASE_URL/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$NEW_REFRESH_TOKEN\"
  }" | jq .
echo ""

# Test 6: Forgot Password
echo -e "${YELLOW}Test 6: Forgot Password${NC}"
curl -s -X POST $BASE_URL/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com"
  }' | jq .
echo ""

# Test 7: Logout
echo -e "${YELLOW}Test 7: Logout${NC}"
curl -s -X POST $BASE_URL/api/v1/auth/logout \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN" | jq .
echo ""

# Test 8: Error Cases
echo -e "${YELLOW}Test 8: Error Case - Wrong Password${NC}"
curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "wrongpassword"
  }' | jq .
echo ""

echo -e "${YELLOW}Test 9: Error Case - Duplicate Registration${NC}"
curl -s -X POST $BASE_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "demo123456",
    "full_name": "Demo User 2",
    "phone": "9876543210"
  }' | jq .
echo ""

echo -e "${YELLOW}Test 10: Error Case - Unauthorized Access${NC}"
curl -s -X GET $BASE_URL/api/v1/user/profile | jq .
echo ""

echo -e "${GREEN}All tests completed!${NC}"
