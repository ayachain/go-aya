#!/bin/bash

go build

mv ./go-aya /usr/local/bin/aya

aya init

cp ./swarm.key /root/.ipfs/swarm.key