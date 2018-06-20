#!/bin/bash
set -e

# Script for creating a lotto environment in fusion

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
vmrun=/Applications/VMware\ Fusion.app/Contents/Library/vmrun
cd ~/Documents/Virtual\ Machines.localized
uname=user
pwd=123
l1="lotto_client_1/lotto_client_1.vmx"
l2="lotto_client_2/lotto_client_2.vmx"
l3="lotto_client_3/lotto_client_3.vmx"
l4="lotto_client_4/lotto_client_4.vmx"

######
# Manual steps
######
# Set up networks
# 1. Create lotto1 with DHCP, Subnet IP: 10.100.0.0, Subnet Mask: 255.255.255.128
lotto1_network_name=$("$vmrun" listHostNetworks | grep 10.100.0.0 | tr -s ' ' | cut -d ' ' -f 2)
sed -i "" -e 's/CHANGEME1/'"$lotto1_network_name"'/g' $script_dir/../fusionvm.json
# 2. Create lotto2 with DHCP, Subnet IP: 10.100.0.128, Subnet Mask: 255.255.255.128
lotto2_network_name=$("$vmrun" listHostNetworks | grep 10.100.0.128 | tr -s ' ' | cut -d ' ' -f 2)
sed -i "" -e 's/CHANGEME2/'"$lotto2_network_name"'/g' $script_dir/../fusionvm.json

#####
# Clean up
#####
printf "#### %s\n" "Cleaning up any existing environment"
"$vmrun" stop $l2 2&>1 > /dev/null || true
"$vmrun" stop $l3 2&>1 > /dev/null || true
"$vmrun" stop $l4 2&>1 > /dev/null || true
rm -r lotto_client_2/* 2&>1 > /dev/null || true
rm -r lotto_client_3/* 2&>1 > /dev/null || true
rm -r lotto_client_4/* 2&>1 > /dev/null || true

######
# Set up client 1
######
# Ensure only 1 interface
printf "#### %s\n" "Setting up client1 before copy, enter password 123 if prompted"
"$vmrun" start $l1 nogui > /dev/null || echo $l1 is already running
# Copy ssh key
until ip=$("$vmrun" getGuestIPAddress $l1)
do
  "$vmrun" -gu $uname -gp $pwd runScriptInGuest $l1 /bin/bash "echo hey"
  ((loops+=1))
  if [ $loops -gt 100 ]; then
    echo could not get ip of client1 exiting
    exit 1
  fi
  sleep 1
done
ssh-copy-id -i ~/.ssh/id_rsa.pub user@$ip
# Install docker
"$vmrun" -gu $uname -gp $pwd runScriptInGuest $l1 /bin/bash 'curl -fsSL get.docker.com -o get-docker.sh; sh get-docker.sh; sudo usermod -aG docker user'
"$vmrun" stop $l1
while [[ $("$vmrun" listNetworkAdapters lotto_client_1/lotto_client_1.vmx | wc -l) -gt 3 ]]; do
  "$vmrun" deleteNetworkAdapter $l1 1
done

######
# Clone
######
printf "#### %s\n" "Cloning client1 to the other clients"
mkdir -p lotto_client_2 lotto_client_3 lotto_client_4
"$vmrun" clone $l1 $l2 linked -cloneName=lotto_client_2
"$vmrun" clone $l1 $l3 linked -cloneName=lotto_client_3
"$vmrun" clone $l1 $l4 linked -cloneName=lotto_client_4

#####
# Hook up to networks
#####
printf "#### %s\n" "Hooking up all clients to networks"
"$vmrun" addNetworkAdapter $l1 custom $lotto1_network_name
"$vmrun" addNetworkAdapter $l2 custom $lotto1_network_name
"$vmrun" addNetworkAdapter $l3 custom $lotto2_network_name
"$vmrun" addNetworkAdapter $l4 custom $lotto2_network_name

####
# Start machines
####
printf "#### %s\n" "Starting all clients"
"$vmrun" start $l1 nogui > /dev/null
"$vmrun" start $l2 nogui > /dev/null
"$vmrun" start $l3 nogui > /dev/null
"$vmrun" start $l4 nogui > /dev/null

#####
# Set ip's
#####
printf "#### %s\n" "Configuring IP settings for clients"
client1_ip=10.100.0.10/25
client2_ip=10.100.0.20/25
client3_ip=10.100.0.150/25
client4_ip=10.100.0.160/25
netplan="network:\n    ethernets:\n        ens38:\n            addresses: [%s]\n"

"$vmrun" -gu $uname -gp $pwd runScriptInGuest $l1 /bin/bash 'printf "'"$netplan"'" "'"$client1_ip"'" | sudo tee /etc/netplan/51-lotto.yaml; sudo netplan apply'
"$vmrun" -gu $uname -gp $pwd runScriptInGuest $l2 /bin/bash 'printf "'"$netplan"'" "'"$client2_ip"'" | sudo tee /etc/netplan/51-lotto.yaml; sudo netplan apply'
"$vmrun" -gu $uname -gp $pwd runScriptInGuest $l3 /bin/bash 'printf "'"$netplan"'" "'"$client3_ip"'" | sudo tee /etc/netplan/51-lotto.yaml; sudo netplan apply'
"$vmrun" -gu $uname -gp $pwd runScriptInGuest $l4 /bin/bash 'printf "'"$netplan"'" "'"$client4_ip"'" | sudo tee /etc/netplan/51-lotto.yaml; sudo netplan apply'

#####
# Success
####
printf "#### %s\n" "Successfully set up"
