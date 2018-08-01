# Basically: nping --tcp-connect -p 80 10.100.0.30

sent=1000
rate=100 # Requests pr second, higher than 5 requires sudo
mode="--tcp-connect"
port=80
# delay=

cmdOut=$(nping -c $sent $mode -p $port --rate $rate 10.100.0.30)
res=$(printf "%s" "$cmdOut" | grep "Successful connections:")

# Possible:
# attempts=$(printf "%s" "$res" | cut -d ' ' -f 4)
# if [ -z $attempts ]; then attempts=0; fi

received=$(printf "%s" "$res" | cut -d ' ' -f 8)
if [ -z $received ]; then received=0; fi
# Or:
# successful=$(printf "%s" "$res" | cut -d ' ' -f 8)
# if [ -z $successful ]; then successful=0; fi

# Possible:
# failed=$(printf "%s" "$res" | cut -d ' ' -f 11)
# if [ -z $failed ]; then failed=0; fi

jq  --arg dataSent $sent \
    --arg dataReceived $received \
    --arg dataRate $rate \
    --arg dataFull "$cmdOut" \
    '. | .["sent"]=($dataSent|tonumber) |
    .["received"]=($dataReceived|tonumber) |
    .["rate"]=($dataRate|tonumber) |
    .["raw"]=$dataFull'<<<'{}'
