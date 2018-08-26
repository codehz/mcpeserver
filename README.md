Minecraft Server Launcher
=========================

A Minecraft Server Launcher Written by Golang.

[![CircleCI](https://circleci.com/gh/codehz/mcpeserver/tree/master.svg?style=svg)](https://circleci.com/gh/codehz/mcpeserver/tree/master)

Powered By [MCMrARM/mcpelauncher-linux](https://github.com/MCMrARM/mcpelauncher-linux).

## Usage

### For ArchLinux

1. Append the repo to `/etc/pacman.conf`
```
[mcpeserver]
SigLevel = Never
Server = https://cdn.codehz.one/repo/archlinux/
```
2. Execute `pacman -Syu mcpeserver mcpeserver-core`
3. Put the minecraft x86 apk to `/srv/mcpeserver`, and then run `cd /srv/mcpeserver && sudo mcpeserver unpack --apk (the apk filename)`
4. Start: `systemctl start mcpeserver@default.service`, Stop: `systemctl stop mcpeserver@default.service`
5. Attach to the server for input command: `mcpeserver attach -profile default`

### For Ubuntu

Comming soon

## Features

* Auto Complete For Command
* Systemd Based Service
* DBus Interface

## LICENSE

GPL v3