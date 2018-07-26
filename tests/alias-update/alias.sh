# alias update script
# Script used for checking that changing the alias of the instance works
set -e

moth="{{.MothershipBinPathAndName}}"
instAlias={{.OriginalAlias}}
ID=$($moth inspect-instance $instAlias -o id)
newAlias=$(echo alias-test-"$(date | shasum | cut -d " " -f 1)")

alias=""
sent=0
received=0

# Update alias 50 times (alternate between instAlias and newAlias)
for i in {1..50}
do
    # Set alias to use
    if (( $i % 2 )); then
        alias=$instAlias
    else
        alias=$newAlias
    fi

    sent=$[$sent + 1]
    # Change alias
    cmdOut=$($moth instance-alias $ID $alias)
    # Verify that the alias was actually changed
    existingAlias=$($moth inspect-instance $ID -o json | jq -r '.alias')
    if [[ "$existingAlias" == "$alias" ]]; then
        received=$[$received + 1]
    fi
    sleep 0.2
done

# Reset to original alias
cmdOut+=$($moth instance-alias $ID $instAlias)

# If none of the commands above failed it means that we were successful
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
