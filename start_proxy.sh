#!/bin/bash

set -e

trap 'kill $(jobs -p)' EXIT # kill all background procs upon exit

make

./distrox config.json 0 &
./distrox config.json 1 &
./distrox config.json 2 &

echo "Started all the nodes!\n"
echo "Press CTRL+C to stop!"
while :
do
    sleep 1
done
