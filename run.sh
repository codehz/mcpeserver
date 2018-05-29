#!/bin/bash
[ ! -e ./games/mcpeserver.sock ] && (./mcpeserver daemon &) &
tail ./games/mcpeserver.log
./mcpeserver attach