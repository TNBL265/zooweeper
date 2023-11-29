#!/bin/bash

if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <num_zooweeper> <num_kafka> <num_react>"
    exit 1
fi

num_zooweeper=$1
num_kafka=$2
num_react=$3
base_url="http://host.docker.internal"

zooweeper_services=""
kafka_services=""
react_services=""

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

# Kafka-Server services
for i in $(seq 0 $(($num_kafka - 1))); do
    port=909$(($i))
    if [ $i -eq 0 ]; then
        build="build:
      context: ./kafka-server
      dockerfile: Dockerfile"
    else
        build=""
    fi
    kafka_services+="
  kafka-server-$i:
    $build
    image: kafka_server:latest
    environment:
      - PORT=$port
      - BASE_URL=$base_url
    ports:
      - \"$port:$port\""
done

# Kafka-React-App services
for i in $(seq 0 $(($num_react - 1))); do
    port=300$(($i))
    if [ $i -eq 0 ]; then
        build="build:
      context: ./kafka-react-app
      dockerfile: Dockerfile"
    else
        build=""
    fi
    react_services+="
  kafka-react-app-$i:
    $build
    image: kafka_react_app:latest
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

$react_services
EOF

echo "docker-compose.yml file generated successfully."

# Run docker-compose up
echo "Starting services with docker-compose up..."
docker-compose up
