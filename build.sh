#!/bin/bash
echo "give me a bottle of rum!"

echo "creating proxy & app"
go build -o proxy cmd/proxy/main.go
go build -o app cmd/app/main.go

echo "creating Docker image app"
mv ./app.D ./Dockerfile
docker build -t app:v1 .
mv ./Dockerfile ./app.D

echo "creating Docker image proxy"
mv ./proxy.D ./Dockerfile
docker build -t proxy:v1 .
mv ./Dockerfile ./proxy.D

echo "docker build the containers"
docker-compose build

echo "run"
docker-compose up

