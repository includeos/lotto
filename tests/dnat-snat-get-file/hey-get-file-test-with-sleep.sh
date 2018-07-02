# hey http testing
# Script used for performing http requests (get file.txt from lotto-client4 10.100.0.160)
# Returns rate and average response time

# Prerequisites to test:
# On lotto-client4 (10.100.0.160):
# - Produce a 1G text file with random content:
# 	base64 /dev/urandom | head -c 1G > file.txt
# - Start apache server on 8080 (returns content of home folder):
# 	docker run -dit --name my-apache-app -p 8080:80 -v "$PWD":/usr/local/apache2/htdocs/ httpd:2.4

# Input values to cmd
sent=1000
concurrency=50

cmdOut=$(docker run --rm rcmorano/docker-hey -n $sent -c $concurrency http://10.100.0.30:1600/file.txt)

# Parse output, important to set a default value if the command over fails
received=$(printf "%s" "$cmdOut" | awk '/responses/ {print $2}' )
if [ -z $received ]; then received=0; fi
rate=$(printf "%s" "$cmdOut" | awk '/Requests\/sec/ {print $2}' )
if [ -z $rate ]; then rate=0; fi
avg=$(printf "%s" "$cmdOut" | awk '/Average/ {print $2}' )
if [ -z $avg ]; then avg=0; fi

jq  --arg dataSent $sent \
	--arg dataReceived $received \
	--arg dataRate $rate \
	--arg dataAvg $avg \
	--arg dataFull "$cmdOut" \
	'. | .["sent"]=($dataSent|tonumber) |
	.["received"]=($dataReceived|tonumber) |
	.["rate"]=($dataRate|tonumber) |
	.["avg"]=($dataAvg|tonumber) |
	.["raw"]=$dataFull'<<<'{}'

sleep 5s
