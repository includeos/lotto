# Basically: nping --udp -p 4242 10.100.0.30 --data-string hi

sent=1000
rate=100 # Requests pr second, higher than 5 requires sudo
mode="--udp"
port=4242
# delay=
data="hi"

cmdOut=$(nping -c $sent --rate $rate $mode -p $port --data-string $data 10.100.0.30)
res=$(printf "%s" "$cmdOut" | grep "UDP packets")

# Possible:
# attempts=$(printf "%s" "$res" | cut -d ' ' -f 4)
# if [ -z $attempts ]; then attempts=0; fi

received=$(printf "%s" "$res" | cut -d ' ' -f 7)
if [ -z $received ]; then received=0; fi
# Or:
# successful=$(printf "%s" "$res" | cut -d ' ' -f 7)
# if [ -z $successful ]; then successful=0; fi

# Possible:
# failed=$(printf "%s" "$res" | cut -d ' ' -f 10)
# if [ -z $failed ]; then failed=0; fi

jq  --arg dataSent $sent \
    --arg dataReceived $received \
    --arg dataRate $rate \
    --arg dataFull "$cmdOut" \
    '. | .["sent"]=($dataSent|tonumber) |
    .["received"]=($dataReceived|tonumber) |
    .["rate"]=($dataRate|tonumber) |
    .["raw"]=$dataFull'<<<'{}'
