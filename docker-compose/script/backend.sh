#!/usr/bin/env bash

while [ ! -f /app/web/dist/index.html ]; do
    echo "Waiting for the web/dist directory..."
    sleep 5
done

GIN_MODE=debug go run main.go migrate
mkdir -p bin/debug
go get -u github.com/cosmtrek/air
GIN_MODE=debug go run github.com/cosmtrek/air