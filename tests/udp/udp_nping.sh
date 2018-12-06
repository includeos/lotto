# Basically: nping --udp -p 4242 10.100.0.30 --data-string hi

sent=1000
rate=100 # Requests pr second, higher than 5 requires sudo
mode="--udp"
port=4242
# delay=
data="hi"

raw=$(nping -c $sent --rate $rate $mode -p $port --data-string $data 10.100.0.30)
res=$(printf "%s" "$raw" | grep "UDP packets")

# Possible:
# attempts=$(printf "%s" "$res" | cut -d ' ' -f 4)
# if [ -z $attempts ]; then attempts=0; fi

received=$(printf "%s" "$res" | cut -d ' ' -f 7)
# Or:
# successful=$(printf "%s" "$res" | cut -d ' ' -f 7)
# if [ -z $successful ]; then successful=0; fi

# Possible:
# failed=$(printf "%s" "$res" | cut -d ' ' -f 10)
# if [ -z $failed ]; then failed=0; fi

if [ "$sent" -eq "$received" ]; then
  success=true
fi

if [ -z $success ]; then success=false; fi
if [ -z $sent ]; then sent=0; fi
if [ -z $received ]; then received=0; fi
if [ -z $rate ]; then rate=0; fi
if [ -z $raw ]; then raw=""; fi
jq \
  --argjson success $success \
  --argjson sent $sent \
  --argjson received $received \
  --argjson rate $rate \
  --arg raw "$raw" \
  '. |
  .["success"]=$success |
  .["sent"]=$sent |
  .["received"]=$received |
  .["rate"]=$rate |
  .["raw"]=$raw
  '<<<'{}'
