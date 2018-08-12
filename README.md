Minecraft Server Launcher
=========================

A Minecraft Server Launcher Written in Golang.

[![CircleCI](https://circleci.com/gh/codehz/mcpeserver/tree/master.svg?style=svg)](https://circleci.com/gh/codehz/mcpeserver/tree/master)

Powered By [MCMrARM/mcpelauncher-linux](https://github.com/MCMrARM/mcpelauncher-linux).

This server software utilizes the built-in server components inside the Minecraft android apk file to run a native Bedrock server. All features are available and you can get Xbox Live achievements on the server, just like realms. Plus more control over the server, it's actually way better than realms.

* Currently the release version supports Minecraft version 1.5.0 and the pre-release version supports 1.5.1.2 as the server core. But all 1.5.x client versions should be able to play on the server.

## Usage

* A Minecraft x86 apk is needed.
* Only specific Minecraft versions are supported.
* I recommend running the server on CentOS 7 or Ubuntu. To run the server on Ubuntu, you have to manually add 32-bit support.

```shell
wget $(curl -s https://api.github.com/repos/codehz/mcpeserver/releases/latest|jq -r '.assets[0].browser_download_url')
chmod +x  mcpeserver
./mcpeserver download # download the core binary for minecraft server
./mcpeserver unpack -apk XXX.apk # unpack assets from minecraft x86 apk
./mcpeserver run # running!
```

Before actually running the server, you might want to edit the server configuration file first.

Server configuration file is located in /games/server.properties for release versions, or /default.cfg for pre-release versions.

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
The preferred way is to put your own world in /games/com.mojang/minecraftWorlds and change the level-dir to the name of your world folder. Otherwise the server will generate a world based on the seed in the config file with some very undesirable settings.

Basic server commands are supported such as list, say, op, etc.
```shell
mcpe://user@localhost.localdomain$ list
There are 1/40 players online:
Cube64128

mcpe://user@localhost.localdomain$ say Hi!
[Server] Hi!

mcpe://user@localhost.localdomain$ op Cube64128
```
To gracefully shut down the server, use the following command:
```shell
mcpe://user@localhost.localdomain$ :quit
```

Mods can be found at https://mcpe.codehz.one/

Refer to [wiki](https://github.com/codehz/mcpeserver/wiki) for other usage.
## Features

* Auto complete of commands.
* Full Minecraft Bedrock server feature support.

## LICENSE

GPL v3
