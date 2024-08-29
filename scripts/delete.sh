#!/bin/bash

# Set the URL for the POST request
url="http://localhost:8000/api/todo/"

id=$1

# Make the POST request using curl
echo "$url$id"
response=$(curl -X DELETE "$url$id")

# Print the response
echo "$response"