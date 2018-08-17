# deploy script
# Script used for checking that deploying an image to an instance (live update) works
set -e

moth="{{.MothershipBinPathAndName}}"
instAlias={{.OriginalAlias}}
instID=$($moth inspect-instance $instAlias -o id)
naclID=$($moth push-nacl tests/deploy/interface.nacl -o id)
# Build an image to deploy to the instance
imgID=$($moth build Starbase --instance $instID --nacl $naclID --tag lotto-deploy-test --waitAndPrint)

sent=0
received=0

# Deploy 100 times
for i in {1..100}
do
    sent=$[$sent + 1]
    # Deploy
    if cmdOut=$($moth deploy $instID $imgID --wait); then
        # Check if the instance now runs the image (note that this will be the same imageId every time):
        cmdOut+=$($moth inspect-instance $instID -o json | jq -r '.imageId')
        if [[ "$cmdOut" == *"$imgID"* ]]; then
            received=$[$received + 1]
        fi
    else
        # Wait up to 1 minute to see if the instance connects back, else finish the test
        for i in {1..60}; do
            sleep 1
            statusNow=$($moth inspect-instance $instID -o json | jq -r '.status')
            if [ "$statusNow" == "connected" ]; then
                # break out of THIS for-loop and continue the test
                break
            else
                # break out of both for-loops (end test)
                break 2
            fi
        done
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
