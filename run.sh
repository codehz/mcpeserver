#!/bin/bash
profile=${1:-default}
[ ! -e ./$profile.sock ] && nohup ./mcpeserver daemon -profile $profile >/dev/null 2>&1 &
while [ ! -e ./$profile.sock ]
do
  sleep 1
done
tail ./$profile.log
./mcpeserver attach