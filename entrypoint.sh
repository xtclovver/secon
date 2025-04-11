#!/bin/sh

# Nginx port is now hardcoded in nginx.conf

echo "Starting backend on port 8081..." # Updated port
# Start the backend application in the background
# Backend now listens on 8081 as configured in config.go and nginx.conf
/app/api &

echo "Starting Nginx on port 8080..." # Nginx port is hardcoded
# Start Nginx in the foreground
nginx -g 'daemon off;'
