# hey http testing (download file via load balancer)
# Script used for performing http requests (get file.txt from lotto-client3 10.100.0.150 and lotto-client4 10.100.0.160)
# Returns rate and average response time

# Prerequisites to test:
# On lotto-client3 (10.100.0.150) and lotto-client4 (10.100.0.160):
# - Produce a 1G text file with random content:
# 	base64 /dev/urandom | head -c 1G > 1GB_file.txt
# - Start apache server on 8080 (returns content of home folder):
#   docker run -dit --name my-apache-app -p 8080:80 -v "$PWD":/usr/local/apache2/htdocs/ httpd:2.4

# Input values to cmd
sent=100
concurrency=50

raw=$(sudo docker run --rm rcmorano/docker-hey -n $sent -c $concurrency http://10.100.0.30:90/1GB_file.txt)

# Parse output, important to set a default value if the command over fails
received=$(printf "%s" "$raw" | awk '/responses/ {print $2}' )
rate=$(printf "%s" "$raw" | awk '/Requests\/sec/ {print $2}' )

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
