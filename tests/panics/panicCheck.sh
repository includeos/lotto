# panicCheck.sh
# Script used for making sure that panic reporting works as intended in Mothership and IncludeOS
set -e

moth="{{.MothershipBinPathAndName}}"
alias={{.OriginalAlias}}
image_ID={{.ImageID}}

# Variables sent as result
sent=1
received=0
rate=1
avg=0

# Check number of panics from before
panics_before=$($moth instance-panics $alias | shasum | cut -d " " -f 1)

# Deploy image to instance
cmdOut="Deploy: ""$($moth deploy --wait $alias $image_ID)"" "

# Wait until the panic has been received
for i in {1..15}; do
  sleep 1
  panics_now=$($moth instance-panics $alias | shasum | cut -d " " -f 1)
  if [ "$panics_now" != "$panics_before" ]; then
    cmdOut+="Time taken to receive panic: $i seconds"
    received=1
    avg=$i
    break
  else
    :
  fi
done

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
