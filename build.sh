#!/bin/bash

# Stop the script if any command fails
set -e

echo "Building the chat app..."

# Ensure that all dependencies are downloaded and the app is built
go mod tidy
go build -o chat-app .