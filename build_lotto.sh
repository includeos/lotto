#!/bin/bash

docker build -t lotto .
docker create --name lot lotto
docker cp lot:/go/bin/lotto lotto
docker rm lot
