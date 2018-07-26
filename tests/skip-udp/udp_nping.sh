# Basically: nping --udp -p 4242 10.100.0.30 --data hei

sent=1000
rate=100 # Requests pr second, higher than 5 requires sudo
mode="--udp"
port=4242
# delay=
data="hi"

cmdOut=$(nping -c $sent $mode -p $port --rate $rate --data $data 10.100.0.30)
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
