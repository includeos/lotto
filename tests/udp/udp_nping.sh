# Basically: nping --udp -p 4242 10.100.0.30 --data-string hi

sent=1000
rate=100 # Requests pr second, higher than 5 requires sudo
mode="--udp"
port=4242
# delay=
data="hi"

cmdOut=$(nping -c $sent --rate $rate $mode -p $port --data-string $data 10.100.0.30)
res=$(printf "%s" "$cmdOut" | grep "Successful connections:")
# attempts=$(printf "%s" "$res" | cut -d ' ' -f 4)
received=$(printf "%s" "$res" | cut -d ' ' -f 7)
# successful=$(printf "%s" "$res" | cut -d ' ' -f 7)
# failed=$(printf "%s" "$res" | cut -d ' ' -f 10)

jq  --arg dataSent $sent \
    --arg dataReceived $received \
    --arg dataRate $rate \
    --arg dataFull "$cmdOut" \
    '. | .["sent"]=($dataSent|tonumber) |
    .["received"]=($dataReceived|tonumber) |
    .["rate"]=($dataRate|tonumber) |
    .["raw"]=$dataFull'<<<'{}'
