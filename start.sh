#!/bin/bash

# Stop the script if any command fails
set -e

# Wait for MySQL to be ready
echo "Waiting for MySQL to be ready..."
until mysql -h "$MYSQL_HOST" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" -e "select 1" > /dev/null 2>&1; do
  echo "MySQL is unavailable - sleeping"
  sleep 3
done

echo "Starting chat application..."

# Start the Go chat application
./chat-app