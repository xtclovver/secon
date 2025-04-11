#!/bin/sh

# Start the backend application in the background
/app/api &

# Start Nginx in the foreground
nginx -g 'daemon off;'
