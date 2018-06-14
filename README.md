# lotto
Long term test platform

## Environments

### Fusion
There are still a few manual steps that must be completed to get fusion to run:

 1. Download client1 from a reliable source.
 2. Place the client1 in a folder named: `~/Documents/Virtual\ Machines.localized/lotto_client_1`
 3. Create 2 custom networks in Vmware Fusion. The names of these should be placed in the `./environments/fusion.sh` script.
 4. Modify `config-mothership.json` to have the correct path to your mothership binary.
 5. Start mothership
