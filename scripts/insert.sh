#!/bin/bash

# Set the URL for the POST request
url="http://localhost:8000/todo"

# Set the data to be sent in the request body
data='{"id":0,"time":"2024-01-01T00:00:00Z", "description": "test", "completed": false}'

# Set the content type header
content_type="Content-Type: application/json"

# Make the POST request using curl
response=$(curl -X POST -H "$content_type" -d "$data" "$url")

# Print the response
echo "$response"