#!/bin/bash

# Start Redis server in the background
redis-server --daemonize yes --dir /data

# Start cron in the background
cron

# Start the Go application
run-app
