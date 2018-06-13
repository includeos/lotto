# Pinger
# Script used for pinging the instance at a fixed rate
# Returns rate and average response time

sent=5
rate=5 # Requests pr second, higher than 5 requires sudo
cmdOut=$(ping -c $sent -i $(awk "BEGIN {print 1/$rate}") -q 10.100.0.200)
received=$(printf "%s" "$cmdOut" | grep received | cut -d ' ' -f 4)
avg=$(printf "%s" "$cmdOut" | grep rtt | grep -oP '=.*?/\K[0-9\.]*')
if [ -z $avg ]; then
  avg=0
fi

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
