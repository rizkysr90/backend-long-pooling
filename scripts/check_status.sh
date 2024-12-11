#!/bin/bash

# Configuration
BASE_URL="http://localhost:8080"  # Replace with your actual base URL
LOG_FILE="job_status_results.log"
INPUT_FILE="curl_results.log"

# Initialize results log file
echo "Timestamp,Job ID,Status Code,Response Time (s),Response Body" > "$LOG_FILE"

# Function to extract job IDs from the input file
extract_job_ids() {
    grep -o '"job_id":"[^"]*"' "$INPUT_FILE" | cut -d'"' -f4
}

# Function to check status for a single job
check_job_status() {
    local job_id=$1
    local request_num=$2
    
    # Capture start time
    start_time=$(date +%s.%N)
    
    # Make the request and capture response
    response=$(curl -s -w "\n%{http_code},%{time_total}" \
        -X GET "$BASE_URL/checkjobstatus/$job_id")
    
    # Split response into body and metrics
    body=$(echo "$response" | head -n 1)
    metrics=$(echo "$response" | tail -n 1)
    
    # Parse metrics
    IFS=',' read -r status_code time_taken <<< "$metrics"
    
    # Get timestamp
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Log the results
    echo "$timestamp,$job_id,$status_code,$time_taken,$body" >> "$LOG_FILE"
    
    # Print progress
    echo "Request $request_num: JobID=$job_id, Status=$status_code, Time=${time_taken}s"
}

echo "Starting status checks for jobs from $INPUT_FILE"
echo "Results will be logged to $LOG_FILE"

# Get all job IDs
job_ids=($(extract_job_ids))
total_jobs=${#job_ids[@]}

echo "Found $total_jobs jobs to check"

# Process jobs in parallel (10 at a time)
for ((i = 0; i < total_jobs; i+=10)); do
    # Launch up to 10 parallel requests
    for ((j = i; j < i+10 && j < total_jobs; j++)); do
        check_job_status "${job_ids[$j]}" "$((j+1))" &
    done
    # Wait for this batch to complete before starting the next
    wait
done

echo "All status checks completed. Results saved in $LOG_FILE"

# Print summary statistics
echo -e "\nSummary Statistics:"
echo "Total Requests: $(wc -l < "$LOG_FILE")"
echo "Success Rate: $(grep ",200," "$LOG_FILE" | wc -l)/$total_jobs"
echo "Average Response Time: $(awk -F',' '{sum+=$4} END {print sum/(NR-1)}' "$LOG_FILE") seconds"

# Show distribution of responses
echo -e "\nResponse Distribution:"
awk -F',' 'NR>1 {print $5}' "$LOG_FILE" | sort | uniq -c | sort -nr