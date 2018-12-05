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

# Check number of panics from before
panics_before=$($moth instance-panics $alias | shasum | cut -d " " -f 1)

# Deploy image to instance
raw="Deploy: ""$($moth deploy --wait $alias $image_ID)"" "

# Wait until the panic has been received
for i in {1..15}; do
  sleep 1
  panics_now=$($moth instance-panics $alias | shasum | cut -d " " -f 1)
  if [ "$panics_now" != "$panics_before" ]; then
    raw+="Time taken to receive panic: $i seconds"
    received=1
    break
  else
    :
  fi
done

if [ "$sent" -eq "$received" ]; then
  result=true
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
