# deploy script
# Script used for checking that deploying an image to an instance (live update) works

moth="{{.MothershipBinPathAndName}}"
instAlias={{.OriginalAlias}}
instID=$($moth inspect-instance $instAlias -o id)
naclID=$($moth push-nacl tests/deploy/interface.nacl {{.BuilderID}} -o id)
# Build an image to deploy to the instance
imgID=$($moth build Starbase {{.BuilderID}} --instance $instID --nacl $naclID --tag lotto-deploy-test --waitAndPrint)

# Deploy 100 times
for i in {1..100}
do
    sent=$[$sent + 1]
    # Deploy
    if raw=$($moth deploy $instID $imgID --wait); then
        # Check if the instance now runs the image (note that this will be the same imageId every time):
        raw+=$($moth inspect-instance $instID -o json | jq -r '.imageId')
        if [[ "$raw" == *"$imgID"* ]]; then
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

if [ "$sent" -eq "$received" ]; then
  success=true
fi

if [ -z $success ]; then success=false; fi
if [ -z $sent ]; then sent=0; fi
if [ -z $received ]; then received=0; fi
if [ -z $rate ]; then rate=0; fi
if [ -z $raw ]; then raw=""; fi
jq \
  --argjson success $success \
  --argjson sent $sent \
  --argjson received $received \
  --argjson rate $rate \
  --arg raw "$raw" \
  '. |
  .["success"]=$success |
  .["sent"]=$sent |
  .["received"]=$received |
  .["rate"]=$rate |
  .["raw"]=$raw
  '<<<'{}'
