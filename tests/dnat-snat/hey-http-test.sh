# Nping arp
# Script used for arp pinging the instance at a fixed rate
# Returns rate and average response time

sent=10000
#rate=100 # Requests pr second
concurrency=200
#cmdOut=$(sudo nping --arp -q --count $sent --rate $rate 10.100.0.30)
cmdOut=$(docker run --rm rcmorano/docker-hey -n $sent -c $concurrency http://10.100.0.30:1500)
#received=$(printf "%s" "$cmdOut" | grep Rcvd | cut -d ' ' -f 8)
received=$(printf "%s" "$cmdOut" | awk '/responses/ {print $2}')
rate=$(printf "%s" "$cmdOut" | awk '/Requests\/sec/ {print $2}')
avg=$(printf "%s" "$cmdOut" | awk '/Average/ {print $2}')

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
