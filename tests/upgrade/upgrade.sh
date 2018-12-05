# upgrade script
# Script used for checking that upgrading an instance works (build image and deploy)
set -e

moth="{{.MothershipBinPathAndName}}"
instAlias={{.OriginalAlias}}
builderID={{.BuilderID}}
instID=$($moth inspect-instance $instAlias -o id)
naclID=$($moth push-nacl tests/upgrade/interface.nacl $builderID -o id)

# 1.
sent=1
received=0
# Upgrade with the pushed nacl
cmdOut=$($moth upgrade $instID --nacl $naclID --waitAndPrintImageID --builderID $builderID)
sleep 0.5
# Check if the instance now reports this new nacl:
cmdOut+=$($moth inspect-instance $instID -o json | jq -r '.naclId')
if [[ "$cmdOut" == *"$naclID"* ]]; then
    received=$[$received + 1]
fi

# 2.
sent=$[$sent + 1]
# Upgrade without specifying nacl (the previously used nacl should be used again)
cmdOut+=$($moth upgrade $instID --waitAndPrintImageID --builderID $builderID)
sleep 0.5
# Check if the instance reports the same nacl:
cmdOut+=$($moth inspect-instance $instID -o json | jq -r '.naclId')
if [[ "$cmdOut" == *"$naclID"* ]]; then
    received=$[$received + 1]
fi

# 3.
sent=$[$sent + 1]
# Upgrade and specify service (Starbase is the only possible service for now)
cmdOut+=$($moth upgrade $instID --service Starbase --waitAndPrintImageID --builderID $builderID)
received=$[$received + 1]

sleep 0.5

# 4.
sent=$[$sent + 1]
customTag="mycustomtag"
# Upgrade and give the image a tag that the instance will report back
cmdOut+=$($moth upgrade $instID --service Starbase --imageTag $customTag --waitAndPrintImageID --builderID $builderID)
sleep 0.5
# Check if the instance now reports this new imageTag:
cmdOut+=$($moth inspect-instance $instID -o json | jq -r '.tag')
if [[ "$cmdOut" == *"$customTag"* ]]; then
    received=$[$received + 1]
fi

# 5.
sent=$[$sent + 1]
# Upgrade and dont specify builder, it should now reuse the same builderID
cmdOut+=$($moth upgrade $instID --service Starbase --waitAndPrintImageID)
received=$[$received + 1]


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
