# Pinger
# Script used for pinging the instance at a fixed rate
# Returns rate and average response time

# Input values to cmd
sent=5
rate=5 # Requests pr second, higher than 5 requires sudo
raw=$(ping -c $sent -i $(awk "BEGIN {print 1/$rate}") -q 10.100.0.30)

# Process output
received=$(printf "%s" "$raw" | grep received | cut -d ' ' -f 4)
if [ "$sent" -eq "$received" ]; then
  result=true
else
  result=false
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
