# Connect to apache server running on client4 (TCP)
# Basically: nping --tcp-connect -p 8080 10.100.0.160

# Prerequisite:
# Run 'docker run -dit --name my-apache-app -p 8080:80 -v "$PWD":/usr/local/apache2/htdocs/ httpd:2.4' on
# lotto-client4 (10.100.0.160)

sent=1000
rate=100 # Requests pr second, higher than 5 requires sudo
mode="--tcp-connect"
port=90
# delay=

cmdOut=$(nping -c $sent $mode -p $port --rate $rate 10.100.0.30)
res=$(printf "%s" "$cmdOut" | grep "Successful connections:")
# attempts=$(printf "%s" "$res" | cut -d ' ' -f 4)
received=$(printf "%s" "$res" | cut -d ' ' -f 8)
# successful=$(printf "%s" "$res" | cut -d ' ' -f 8)
# failed=$(printf "%s" "$res" | cut -d ' ' -f 11)

jq  --arg dataSent $sent \
    --arg dataReceived $received \
    --arg dataRate $rate \
    --arg dataFull "$cmdOut" \
    '. | .["sent"]=($dataSent|tonumber) |
    .["received"]=($dataReceived|tonumber) |
    .["rate"]=($dataRate|tonumber) |
    .["raw"]=$dataFull'<<<'{}'
