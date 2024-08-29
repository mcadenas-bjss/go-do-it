#!/bin/bash

id=$1

# Set the URL for the POST request
url="http://localhost:8000/api/todo/"

description=$2

# Set the data to be sent in the request body
data='{"id":0,"time":"2024-01-01T00:00:00Z", "description": "'$description'", "completed": false}'

echo $data

# Set the content type header
content_type="Content-Type: application/json"

# Make the POST request using curl
response=$(curl -X PUT -H "$content_type" -d "$data" "$url$id")

# Print the response
echo "$response"