# gophertunnel
A Minecraft library containing packages to create clients, servers, proxies and other tools, and a proxy implementation using them.

[Module Documentation](https://pkg.go.dev/mod/github.com/sandertv/gophertunnel)

![telescope gopher](https://github.com/Sandertv/gophertunnel/blob/master/gophertunnel_telescope_coloured.png)

## Overview
gophertunnel is composed of several packages that may be of use for creating Minecraft related tools. A brief
overview of all packages may be found [here](https://pkg.go.dev/mod/github.com/sandertv/gophertunnel?tab=packages).

## Examples


## Versions
Gophertunnel supports only one version at a time. Generally, a new minor version is tagged when gophertunnel
supports a new Minecraft version that was not previously supported. A list of the recommended gophertunnel
versions for past Minecraft versions is listed below.

| Version | Tag      |
|---------|----------|
| 1.16.20 | Latest   |
| 1.16.0  | v1.7.11  |
| 1.14.60 | v1.6.5   |
| 1.14.0  | v1.3.20  |
| 1.13.0  | v1.3.5   |
| 1.12.0  | v1.2.11  |

## Proxy
A MITM proxy program is implemented in the main.go file. It uses the gophertunnel libraries to create a proxy
that provides user authentication and proxying a connection to another server.

## Sponsors
Gophertunnel is sponsored by all my [gopher patrons](https://patreon.com/sandertv). A special thanks goes to
the Very Important Gophers below.
<a href="https://github.com/TwistedAsylumMC"><img src="https://avatars3.githubusercontent.com/u/30378179?s=400&u=49eabab31601b6bf5b0024c05c2556bc7f5b3e3b&v=4" width="60" height="60"></a>

## Contact
[![Chat on Discord](https://img.shields.io/badge/Chat-On%20Discord-738BD7.svg?style=for-the-badge)](https://discord.gg/evzQR4R)