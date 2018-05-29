#!/bin/bash
[ ! -e ./games/mcpeserver.sock ] && (./mcpeserver daemon &) &
while [ ! -e ./games/mcpeserver.sock ]
do
  sleep 1
done
tail ./games/mcpeserver.log
./mcpeserver attach