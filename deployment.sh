#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <num_server>"
    exit 1
fi

num_zooweeper=$1

num_kafka=3
num_frontend=2
base_url="http://host.docker.internal"

zooweeper_services=""
kafka_services=""
frontend_services=""

# ZooWeeper services
start_port=8080
end_port=$(($start_port + $num_zooweeper - 1))
for i in $(seq 0 $(($num_zooweeper - 1))); do
    port=$(($start_port + $i))
    if [ $i -eq 0 ]; then
        build="build:
      context: ./server
      dockerfile: Dockerfile"
    else
        build=""
    fi
    zooweeper_services+="
  zooweeper-$i:
    $build
    image: zooweeper_server:latest
    environment:
      - PORT=$port
      - START_PORT=$start_port
      - END_PORT=$end_port
      - BASE_URL=$base_url
    ports:
      - \"$port:$port\""
done

# Kafka broker services
for i in $(seq 0 $(($num_kafka - 1))); do
    port=909$(($i))
    if [ $i -eq 0 ]; then
        build="build:
      context: ./kafka-broker
      dockerfile: Dockerfile"
    else
        build=""
    fi
    kafka_services+="
  kafka-broker-$i:
    $build
    image: kafka_broker:latest
    environment:
      - PORT=$port
      - BASE_URL=$base_url
    ports:
      - \"$port:$port\""
done

# Frontend services
for i in $(seq 0 $(($num_frontend - 1))); do
    port=300$(($i))
    if [ $i -eq 0 ]; then
        build="build:
      context: ./frontend
      dockerfile: Dockerfile"
    else
        build=""
    fi
    frontend_services+="
  frontend-$i:
    $build
    image: frontend:latest
    environment:
      - PORT=$port
      - BASE_URL=$base_url
    ports:
      - \"$port:$port\""
done

# Combine services into docker-compose.yml
cat <<EOF > docker-compose.yml
version: "3.9"

services:
$zooweeper_services

$kafka_services

$frontend_services
EOF

echo "docker-compose.yml file generated successfully."

# Run docker-compose up
echo "Starting services with docker-compose up..."
docker-compose up
