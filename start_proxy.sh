#!/bin/bash

set -e
trap 'kill $(jobs -p)' EXIT # kill all background procs upon exit

go run proxy_node.go config.json 0 &
go run proxy_node.go config.json 1 &
go run proxy_node.go config.json 2 &

echo "Started all the nodes!\n"
echo "Press CTRL+C to stop!"
while :
do
    sleep 1
done
