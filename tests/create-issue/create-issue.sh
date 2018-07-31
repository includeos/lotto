# create issue script
# Script used for checking that it is possible to create an issue
set -e

moth="{{.MothershipBinPathAndName}}"

sent=0
received=0

# create-issue --name <name> --type <type> --description <description>
# delete-issue
# inspect-issue
# issues
# issuetypes
# pull-issue

issueNameBase="lotto-issue"

# Create an issue 3 times and verify that the issue was created
for i in {1..3}
do
    issueName="$issueNameBase-$i"
    sent=$[$sent + 1]
    # Create issue
    createdIssueID=$($moth create-issue --name $issueName --type Deployment --description "This is an issue created by Lotto" -o id)
    # Verify that the issue was actually created
    nameOfIssueCreated=$($moth inspect-issue $createdIssueID -o json | jq -r '.name')
    if [[ "$nameOfIssueCreated" == "$issueName" ]]; then
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
