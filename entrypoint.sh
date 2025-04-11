#!/bin/sh

# Set default port if PORT environment variable is not set
export PORT=${PORT:-8080}

# Substitute environment variables in the Nginx config template
# We copy the original config to a template location first,
# then use envsubst to create the final config.
# This assumes nginx.conf was copied to /etc/nginx/nginx.conf.template in Dockerfile
# Let's adjust this: we'll modify the file in place after copying it.
# A safer approach is to use a template file. Let's modify nginx.conf directly for simplicity now,
# but using a template is better practice.

# Substitute the PORT variable in nginx.conf
# We need envsubst for this, which requires the 'gettext' package.
# We'll add 'gettext' installation to the Dockerfile.
envsubst '${PORT}' < /etc/nginx/nginx.conf > /etc/nginx/nginx.conf.temp && mv /etc/nginx/nginx.conf.temp /etc/nginx/nginx.conf

echo "Starting backend on port 8080..."
# Start the backend application in the background
# Assuming the backend listens on 8080 as per nginx proxy_pass
/app/api &

echo "Starting Nginx on port ${PORT}..."
# Start Nginx in the foreground
nginx -g 'daemon off;'
