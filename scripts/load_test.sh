#!/bin/bash

# Configuration
ENDPOINT="http://localhost:8080/insert"  # Replace with your actual endpoint
NUM_REQUESTS=100
LOG_FILE="curl_results.log"

# Initialize log file
echo "Timestamp,Request#,Status,Response Time (s),Job ID" > "$LOG_FILE"

# Function to make a single request and log results
make_request() {
    local request_num=$1
    
    # Capture start time
    start_time=$(date +%s.%N)
    
    # Make the request and capture response
    response=$(curl -s -w "\n%{http_code},%{time_total}" -X POST "$ENDPOINT")
    
    # Split response into body and metrics
    body=$(echo "$response" | head -n 1)
    metrics=$(echo "$response" | tail -n 1)
    
    # Parse metrics
    IFS=',' read -r status_code time_taken <<< "$metrics"
    
    # Get timestamp
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Log the results
    echo "$timestamp,$request_num,$status_code,$time_taken,$body" >> "$LOG_FILE"
    
    # Print progress
    echo "Request $request_num: Status=$status_code, Time=${time_taken}s, JobID=$body"
}

echo "Starting $NUM_REQUESTS POST requests to $ENDPOINT"
echo "Results will be logged to $LOG_FILE"

# Make requests in parallel (10 at a time)
for ((i = 1; i <= NUM_REQUESTS; i+=10)); do
    # Launch up to 10 parallel requests
    for ((j = i; j < i+10 && j <= NUM_REQUESTS; j++)); do
        make_request $j &
    done
    # Wait for this batch to complete before starting the next
    wait
done

echo "All requests completed. Results saved in $LOG_FILE"

# Print summary statistics
echo -e "\nSummary Statistics:"
echo "Total Requests: $(wc -l < "$LOG_FILE")"
echo "Success Rate: $(grep ",200," "$LOG_FILE" | wc -l)/$NUM_REQUESTS"
echo "Average Response Time: $(awk -F',' '{sum+=$4} END {print sum/(NR-1)}' "$LOG_FILE") seconds"