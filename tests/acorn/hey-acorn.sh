# hey acorn http testing
# Script used for performing http requests
# Returns rate and average response time

# Input values to cmd
sent=1000
concurrency=100
raw=$(docker run --rm rcmorano/docker-hey -n $sent -c $concurrency http://10.100.0.30)

# Parse output, important to set a default value if the command over fails
receivedStatus=$(printf "%s" "$raw" | awk '/responses/ {print $1}' )
received=$(printf "%s" "$raw" | awk '/responses/ {print $2}' )

if [[ "$receivedStatus" != "[200]" ]]; then
    received=0;
elif [ -z $received ]; then
    received=0;
fi

if [ "$sent" -eq "$received" ]; then
  success=true
fi

rate=$(printf "%s" "$raw" | awk '/Requests\/sec/ {print $2}' )

if [ -z $success ]; then success=false; fi
if [ -z $sent ]; then sent=0; fi
if [ -z $received ]; then received=0; fi
if [ -z $rate ]; then rate=0; fi
if [ -z $raw ]; then raw=""; fi
jq \
  --argjson success $success \
  --argjson sent $sent \
  --argjson received $received \
  --argjson rate $rate \
  --arg raw "$raw" \
  '. |
  .["success"]=$success |
  .["sent"]=$sent |
  .["received"]=$received |
  .["rate"]=$rate |
  .["raw"]=$raw
  '<<<'{}'
