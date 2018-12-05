# Pinging 10.100.0.160
# Script used for pinging the instance at a fixed rate
# Returns rate and average response time

sent=20
rate=5 # Requests pr second, higher than 5 requires sudo
cmdOut=$(ping -c $sent -i $(awk "BEGIN {print 1/$rate}") -q 10.100.0.160)
received=$(printf "%s" "$cmdOut" | grep received | cut -d ' ' -f 4)

# Fail test if there were received packets. This test should not succeed
if [ "$received" -eq "0" ]; then
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
