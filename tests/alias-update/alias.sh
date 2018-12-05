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
    raw=$($moth instance-alias $ID $alias)
    # Verify that the alias was actually changed
    existingAlias=$($moth inspect-instance $ID -o json | jq -r '.alias')
    if [[ "$existingAlias" == "$alias" ]]; then
        received=$[$received + 1]
    fi
    sleep 0.2
done

# Reset to original alias
raw+=$($moth instance-alias $ID $instAlias)

# All commands above succeeded
result=true

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
