#!/bin/bash

# Compile Go application
go build -o app

# Make sure templates directory exists in the right locations
mkdir -p /templates
cp -r templates/* /templates/

# For debugging
echo "Contents of current directory:"
ls -la

echo "Contents of /templates directory:"
ls -la /templates

# Set execution permission
chmod +x app
