# hey http testing
# Script used for performing http requests
# Returns rate and average response time

# Prerequesites:
# Requires a web server to run on client3

# Input values to cmd
sent=1000
concurrency=200
timeout=2
raw=$(docker run --rm rcmorano/docker-hey -t $timeout -n $sent -c $concurrency http://10.100.0.30:1500)

# Parse output, important to set a default value if the command over fails
received=$(printf "%s" "$raw" | awk '/responses/ {print $2}' )
rate=$(printf "%s" "$raw" | awk '/Requests\/sec/ {print $2}' )

# Only passes if 100% of packets were received
if [ "$sent" -eq "$received" ]; then
  result=true
else
  result=false
fi

if [ -z $result ]; then result=false; fi
if [ -z $sent ]; then sent=0; fi
if [ -z $received ]; then received=0; fi
if [ -z $rate ]; then rate=0; fi
if [ -z $raw ]; then raw=""; fi
jq \
  --argjson result $result \
  --argjson sent $sent \
  --argjson received $received \
  --argjson rate $rate \
  --arg raw "$raw" \
  '. |
  .["result"]=$result |
  .["sent"]=$sent |
  .["received"]=$received |
  .["rate"]=$rate |
  .["raw"]=$raw
  '<<<'{}'
