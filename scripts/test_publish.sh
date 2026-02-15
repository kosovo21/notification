#!/bin/bash

API_KEY="test-api-key"
URL="http://localhost:8080/api/v1/messages/send"

echo "Sending message..."
curl -v -X POST $URL \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "to": ["+1234567890", "user@example.com"],
    "from": "TestSender",
    "message": "Hello from RabbitMQ!",
    "platform": "sms",
    "priority": 1,
    "subject": "Test Message"
  }'

echo -e "\n\nChecking message list..."
curl -v -X GET "http://localhost:8080/api/v1/messages" \
  -H "X-API-Key: $API_KEY"
