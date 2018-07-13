Minecraft Server Launcher
=========================

A Minecraft Server Launcher Written by Golang.

[![CircleCI](https://circleci.com/gh/codehz/mcpeserver/tree/master.svg?style=svg)](https://circleci.com/gh/codehz/mcpeserver/tree/master)

Powered By [MCMrARM/mcpelauncher-linux](https://github.com/MCMrARM/mcpelauncher-linux).

## Usage

```shell
wget $(curl -s https://api.github.com/repos/codehz/mcpeserver/releases/latest|jq -r '.assets[0].browser_download_url')
chmod +x  mcpeserver
./mcpeserver download # download the core binary for minecraft server
./mcpeserver unpack -apk XXX.apk # unpack assets from minecraft
./mcpeserver run # running!
```

* You must provide a valid minecraft x86 apk

More document can be found in [wiki](https://github.com/codehz/mcpeserver/wiki)

## Features

* Auto Complete For Command

## LICENSE

GPL v3