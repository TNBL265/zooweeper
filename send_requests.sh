#!/bin/bash

if [ -z "$1" ]; then
  echo "Usage: $0 <num_requests> <kafka-port> <other-clients>"
  exit 1
fi

num_requests=$1

# Kafka broker random choices
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

# Record the start time with nanoseconds before the loop
start_time=$(date +%s.%N)

# Loop for sending requests
minute=1

for i in $(seq 1 "$num_requests"); do
  # ZooWeeper Server random choices
  receiver_ip=$((RANDOM % 3 + 8080))
  name=${names[$RANDOM % ${#names[@]}]}
  club=${clubs[$RANDOM % ${#clubs[@]}]}
  score=$(generate_score)

  # Use curl to send the request and capture the response
  response=$(curl -s -X POST "$base_url" \
       -H "Content-Type: application/json" \
       -d "$(cat <<EOF
{
    "metadata": {
      "ReceiverIp": "$receiver_ip",
      "Clients": "$clients"
    },
    "gameResults": {
        "Minute": $minute,
        "Player": "$name",
        "Club": "$club",
        "Score": "$score"
    }
}
EOF
)")

  # Check if the response contains "OK"
  if echo "$response" | grep -q "OK"; then
    echo "Received OK response for request $i."
  else
    echo "Received unexpected response for request $i: $response"
  fi

  minute=$((minute + 1))
  if [ $minute -gt 90 ]; then
    minute=1
  fi
done

# Record the end time with nanoseconds after the last request
end_time=$(date +%s.%N)

# Calculate the total time taken with nanoseconds precision using awk
total_time_seconds=$(awk "BEGIN {print $end_time - $start_time}")

echo "Total time taken for $num_requests requests: $total_time_seconds seconds"
echo "Requests sent."
