#!/bin/bash

docker network create distributed-network || 'true'
docker volume create db-data || 'true'
docker compose up -d --build