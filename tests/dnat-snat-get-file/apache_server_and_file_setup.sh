if [ ! -f 1GB_file.txt ]; then
  base64 /dev/urandom | head -c 1G > 1GB_file.txt
fi
docker run --rm -dit --name my-apache-app -p 8080:80 -v "$PWD":/usr/local/apache2/htdocs/ httpd:2.4
