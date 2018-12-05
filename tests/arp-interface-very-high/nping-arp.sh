# Nping arp
# Script used for arp pinging the instance at a fixed rate
# Returns rate and average response time

# Input values to cmd
sent=500
target=450
rate=250 # Requests pr second, higher than 5 requires sudo
raw=$(sudo nping --arp -q --count $sent --rate $rate 10.100.0.30)

# Parse output
received=$(printf "%s" "$raw" | grep Rcvd | cut -d ' ' -f 8)

# If we receive more than 450 then the test passes
if [ "$received" -gt "$target" ]; then
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
