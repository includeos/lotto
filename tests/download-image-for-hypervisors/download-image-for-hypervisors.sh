# download image script
# Script used for checking that downloading an image for each hypervisor works
set -e

moth="{{.MothershipBinPathAndName}}"

# Setup: Build an image to download
instAlias={{.OriginalAlias}}
instID=$($moth inspect-instance $instAlias -o id)
naclID=$($moth push-nacl tests/download-image-for-hypervisors/interface.nacl {{.BuilderID}} -o id)
imgID=$($moth build Starbase {{.BuilderID}} --instance $instID --nacl $naclID --waitAndPrint)

hypervisor=""
sent=0
received=0

# Download the image for hypervisors qemu, vcloud and virtualbox
for i in {1..3}
do
    if [ $i == 1 ]; then
        hypervisor="vcloud"
    elif [ $i == 2 ]; then
        hypervisor="virtualbox"
    else
        hypervisor="qemu"
    fi
    downloadedImgName="img-$hypervisor"
    sent=$[$sent + 1]
    # Download image
    cmdOut=$($moth pull-image $imgID $downloadedImgName --format $hypervisor)
    received=$[$received + 1]
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
