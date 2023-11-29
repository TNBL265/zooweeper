#!/bin/bash

if [ -z "$1" ]; then
  echo "Usage: $0 <num_requests> <kafka-port> <other-clients>"
  exit 1
fi

num_requests=$1

# Kafka Server random choices
ports=(9090 9091 9092)
selected_port=${ports[$RANDOM % ${#ports[@]}]}

kafka_port=${2:-$selected_port}
clients=${3:-"9090,9091,9092"}

base_url="http://localhost:$kafka_port/addScore"

names=("Ronaldo" "Messi" "Pele" "Maradona")
clubs=("FCB" "RMA" "MU" "FCB")

# Generate a random score
generate_score() {
  local team1=$((RANDOM % 5))
  local team2=$((RANDOM % 5))
  echo "${team1}-${team2}"
}

send_post_request() {
  local ip=$1
  local minute=$2
  local player=$3
  local club=$4
  local score=$5

  local payload=$(cat <<EOF
{
    "metadata": {
      "ReceiverIp": "$ip",
      "Clients": "$clients"
    },
    "gameResults": {
        "Minute": $minute,
        "Player": "$player",
        "Club": "$club",
        "Score": "$score"
    }
}
EOF
)

  # Use curl to send a POST request
  curl -X POST "$base_url" \
       -H "Content-Type: application/json" \
       -d "$payload"
}

minute=1

for i in $(seq 1 "$num_requests"); do
  # ZooWeeper Server random choices
  receiver_ip=$((RANDOM % 3 + 8080))
  name=${names[$RANDOM % ${#names[@]}]}
  club=${clubs[$RANDOM % ${#clubs[@]}]}
  score=$(generate_score)

  send_post_request "$receiver_ip" $minute "$name" "$club" "$score"

  minute=$((minute + 1))
  if [ $minute -gt 90 ]; then
    minute=1
  fi
done

echo "Requests sent."
