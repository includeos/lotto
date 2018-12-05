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
    raw+=$($moth pull-image $imgID $downloadedImgName --format $hypervisor 2>&1)
    if [ "$?" -eq "0" ]; then
      received=$[$received + 1]
    fi
done

# If none of the commands above failed it means that we were successful
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
