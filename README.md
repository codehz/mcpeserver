Minecraft Server Launcher
=========================

A Minecraft Server Launcher Written in Golang.

[![CircleCI](https://circleci.com/gh/codehz/mcpeserver/tree/master.svg?style=svg)](https://circleci.com/gh/codehz/mcpeserver/tree/master)

Powered By [MCMrARM/mcpelauncher-linux](https://github.com/MCMrARM/mcpelauncher-linux).

This server software utilizes the built-in server components inside the Minecraft android apk file to run a native Bedrock server. All features are available and you can get Xbox Live achievements on the server, just like realms. Plus more control over the server, it's actually way better than realms.

* Currently the release version supports Minecraft version 1.5.0 and the pre-release version supports 1.5.1.2 as the server core. But all 1.5.x client versions should be able to play on the server.

## Features

* Auto Complete For Command
* Full Minecraft Bedrock server feature support.
* Systemd Based Service
* DBus Interface

## Installation

### For ArchLinux

1. Append the repo to `/etc/pacman.conf`
```
[mcpeserver]
SigLevel = Never
Server = https://cdn.codehz.one/repo/archlinux/
```
2. Execute `pacman -Syu mcpeserver mcpeserver-core`
3. Execute `systemctl reload dbus`
4. Put the minecraft x86 apk to `/srv/mcpeserver`, and then run `cd /srv/mcpeserver && sudo mcpeserver unpack --apk (the apk filename)`
5. Start: `systemctl start mcpeserver@default.service`, Stop: `systemctl stop mcpeserver@default.service`
6. Attach to the server for input command: `mcpeserver attach -profile default`

### For Ubuntu

Comming soon

## Usage

You might want to edit the server configuration file before actually running the server.

Server configuration file is located in /srv/micpeserver/default.cfg.

Here is an example of the server configuration file.
```shell
level-dir=world
level-name=§aServer example
level-generator=1
level-seed=1019130957
difficulty=3
gamemode=0
force-gamemode=false
motd=§6Welcome to §9server example!
server-port=19132
server-port-v6=19133
max-players=40
online-mode=true
view-distance=56
player-idle-timeout=0
```
The preferred way is to put your own world in /srv/mcpeserver/worlds and change the level-dir to the name of your world folder. Otherwise the server will generate a world based on the seed in the config file with some very undesirable settings.

(Tips: make sure all files in /srv/mcpeserver can be accessed by mcpeserver user)

Basic server commands are supported such as list, say, op, etc.
```shell
socket://user@localhost.localdomain$ /list
There are 1/40 players online:
CodeHz

socket://user@localhost.localdomain$ /say Hi!
[Server] Hi!

socket://user@localhost.localdomain$ /op CodeHz
```

Refer to [wiki](https://github.com/codehz/mcpeserver/wiki) for other usage.

## LICENSE

GPL v3
