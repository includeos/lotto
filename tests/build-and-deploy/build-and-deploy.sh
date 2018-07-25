# build and deploy script
# Script used for checking that building an image and deploying it to an instance (live update) works
set -e

moth="{{.MothershipBinPathAndName}}"
instAlias={{.OriginalAlias}}
instID=$($moth inspect-instance $instAlias -o id)
naclID=$($moth push-nacl tests/deploy/interface.nacl -o id)
tagBase="image"

sent=0
received=0

# Build and deploy 20 times
for i in {1..20}
do
    tag="$tagBase-$i"
    sent=$[$sent + 1]
    # Build
    imgID=$($moth build Starbase --instance $instID --nacl $naclID --tag $tag --waitAndPrint)
    # Deploy
    cmdOut+=$($moth deploy $instID $imgID --wait)
    # Check if the instance now reports this new nacl:
    cmdOutImgID=$($moth inspect-instance $instID -o json | jq -r '.imageId')
    cmdOutTag=$($moth inspect-instance $instID -o json | jq -r '.tag')
    if [[ "$cmdOutImgID" == *"$imgID"* ]] && [[ "$cmdOutTag" == *"$tag"* ]]; then
        received=$[$received + 1]
    fi
done

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
