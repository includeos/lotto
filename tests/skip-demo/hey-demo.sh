# hey demo (service) http testing
# Script used for performing http requests
# Returns rate and average response time

# Input values to cmd
sent=10000
concurrency=200
cmdOut=$(docker run --rm rcmorano/docker-hey -n $sent -c $concurrency http://10.100.0.30)

# Parse output, important to set a default value if the command over fails
received=$(printf "%s" "$cmdOut" | awk '/responses/ {print $2}' )
if [ -z $received ]; then received=0; fi
rate=$(printf "%s" "$cmdOut" | awk '/Requests\/sec/ {print $2}' )
if [ -z $rate ]; then rate=0; fi
avg=$(printf "%s" "$cmdOut" | awk '/Average/ {print $2}' )
if [ -z $avg ]; then avg=0; fi

jq  --arg dataSent $sent \
    --arg dataReceived $received \
    --arg dataRate $rate \
    --arg dataAvg $avg \
    --arg dataFull "$cmdOut" \
    '. | .["sent"]=($dataSent|tonumber) |
    .["received"]=($dataReceived|tonumber) |
    .["rate"]=($dataRate|tonumber) |
    .["avg"]=($dataAvg|tonumber) |
    .["raw"]=$dataFull'<<<'{}'
