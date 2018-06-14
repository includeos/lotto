#!/bin/bash
set -e

# Script for creating a lotto environment in fusion

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
lotto1_network_name=vmnet5 #changing to a custom name does not work
# 2. Create lotto2 with DHCP, Subnet IP: 10.100.0.128, Subnet Mask: 255.255.255.128
lotto2_network_name=vmnet6 #changing to a custom name does not work
# 3. Add SSH key to .ssh/authorized keys for client1

#####
# Clean up
#####
printf "#### %s\n" "Cleaning up any existing environment"
"$vmrun" stop $l2 || echo $l2 is not running
"$vmrun" stop $l3 || echo $l3 is not running
"$vmrun" stop $l4 || echo $l4 is not running
rm -r lotto_client_2/* || echo lotto_client_2 folder does not need to be deleted
rm -r lotto_client_3/* || echo lotto_client_3 folder does not need to be deleted
rm -r lotto_client_4/* || echo lotto_client_4 folder does not need to be deleted

######
# Set up client 1
######
# Ensure only 1 interface
printf "#### %s\n" "Setting up client1 before copy"
"$vmrun" start $l1 || echo $l1 does not need to be started
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
