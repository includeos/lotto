# Connect to apache server running on client4 (TCP)
# Basically: nping --tcp-connect -p 1600 10.100.0.30
# Should be dnat'ed to 10.100.0.160 port 8080

# Prerequisite:
# Run 'docker run -dit --name my-apache-app -p 8080:80 -v "$PWD":/usr/local/apache2/htdocs/ httpd:2.4' on
# lotto-client4 (10.100.0.160)

sent=1000
rate=200 # Requests pr second, higher than 5 requires sudo
mode="--tcp-connect"
port="1600"
# delay=

raw=$(nping -c $sent $mode -p $port --rate $rate 10.100.0.30)

res=$(printf "%s" "$raw" | grep "Successful connections:")

# attempts=$(printf "%s" "$res" | cut -d ' ' -f 4)
received=$(printf "%s" "$res" | cut -d ' ' -f 8)
# successful=$(printf "%s" "$res" | cut -d ' ' -f 8)
# failed=$(printf "%s" "$res" | cut -d ' ' -f 11)

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
