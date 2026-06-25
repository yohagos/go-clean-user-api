#!/bin/bash

if ! command -v swag &> /dev/null
then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi


echo "Generating Swagger documentation.."
swag init -g cmd/api/main.go -o docs

echo "Swagger documentation generated successfully!"
echo "Run 'docker-compose up' and visit http://localhost:8080/swagger/index.html"