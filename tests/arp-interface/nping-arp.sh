# Nping arp
# Script used for arp pinging the instance at a fixed rate
# Returns rate and average response time

sent=500
rate=100 # Requests pr second, higher than 5 requires sudo
cmdOut=$(sudo nping --arp -q --count $sent --rate $rate 10.100.0.30)
received=$(printf "%s" "$cmdOut" | grep Rcvd | cut -d ' ' -f 8)
avg=0

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
