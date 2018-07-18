# alias update script
# Script used for checking that changing the alias of the instance works
set -e

moth="{{.MothershipBinPathAndName}}"
original_alias={{.OriginalAlias}}
ID=$($moth inspect-instance $original_alias -o id)
new_alias=$(echo alias-test-"$(date | shasum | cut -d " " -f 1)")

# Change alias to new_alias
cmdOut=$($moth instance-alias $ID $new_alias)

sleep 0.2

# Change alias back to old alias
cmdOut+=$($moth instance-alias $ID $original_alias)

# If none of the commands above failed it means that we were successful
sent=1
received=1
rate=0.1
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
