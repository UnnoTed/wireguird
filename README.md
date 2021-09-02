# wireguird

##### a linux gtk gui client for [Wireguard](https://www.wireguard.com/)

________________
Features:

- System tray icon goes red when connected, black when disconnected.
- Looks the same and does almost the same things as the official Wireguard's Windows gui client.
- Lists tunnels from `/etc/wireguard`
- Controls Wireguard ~~*through*~~ `wg-quick`

## Preview (video)

[![wireguird preview](https://raw.githubusercontent.com/UnnoTed/wireguird/master/preview.png)](https://streamable.com/dpthpr)

## Download

##### Ubuntu

tested on: `18.04 LTS`, `20.04 LTS` and `21.04`

[wireguird_amd64.deb (1.8mb)](https://github.com/UnnoTed/wireguird/releases/download/v0.2.0/wireguird_amd64.deb)

## Compile

dependencies: `wireguard-tools libgtk-3-dev libappindicator3-dev`

```sh
git clone https://github.com/UnnoTed/wireguird
cd wireguird
./deps.sh
./package.sh
./install.sh
```
