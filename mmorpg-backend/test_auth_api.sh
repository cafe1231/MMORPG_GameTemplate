#!/bin/bash

# Test Auth Service API endpoints

echo "=== Testing Auth Service API ==="
AUTH_URL="http://localhost:8081/api/v1/auth"
GATEWAY_URL="http://localhost:8090/api/v1/auth"

# Function to test endpoint
test_endpoint() {
    local url=$1
    local method=$2
    local data=$3
    local desc=$4
    
    echo -e "\n--- Testing: $desc ---"
    echo "URL: $url"
    echo "Method: $method"
    if [ -n "$data" ]; then
        echo "Data: $data"
        curl -X $method -H "Content-Type: application/json" -d "$data" $url -w "\nHTTP Status: %{http_code}\n"
    else
        curl -X $method $url -w "\nHTTP Status: %{http_code}\n"
    fi
}

# Test direct auth service
echo -e "\n=== Testing Direct Auth Service (port 8081) ==="

# 1. Health check
test_endpoint "http://localhost:8081/health" "GET" "" "Health Check"

# 2. Register new user
REGISTER_DATA='{
    "email": "test@example.com",
    "password": "TestPass123!",
    "username": "testuser",
    "acceptTerms": true
}'
test_endpoint "$AUTH_URL/register" "POST" "$REGISTER_DATA" "User Registration"

# 3. Login
LOGIN_DATA='{
    "email": "test@example.com",
    "password": "TestPass123!"
}'
test_endpoint "$AUTH_URL/login" "POST" "$LOGIN_DATA" "User Login"

# Test via Gateway
echo -e "\n\n=== Testing via Gateway (port 8090) ==="

# 1. Gateway health check
test_endpoint "http://localhost:8090/health" "GET" "" "Gateway Health Check"

# 2. Register via gateway
test_endpoint "$GATEWAY_URL/register" "POST" "$REGISTER_DATA" "Register via Gateway"

# 3. Login via gateway
test_endpoint "$GATEWAY_URL/login" "POST" "$LOGIN_DATA" "Login via Gateway"

echo -e "\n=== Test Complete ==="