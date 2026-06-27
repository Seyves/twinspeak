#!/bin/sh
set -e

# Set defaults if not provided
export BACKEND_HOST=${BACKEND_HOST:-backend:8080}
export FRONTEND_HOST=${FRONTEND_HOST:-frontend:4321}

envsubst '${BACKEND_HOST} ${FRONTEND_HOST}' < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf

nginx -g 'daemon off;'
