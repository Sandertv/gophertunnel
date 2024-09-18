# gophertunnel
> Swiss army knife for Minecraft (Bedrock Edition) software written in Go

[![PkgGoDev](https://pkg.go.dev/badge/github.com/sandertv/gophertunnel)](https://pkg.go.dev/github.com/sandertv/gophertunnel)

![telescope gopher](https://raw.githubusercontent.com/Sandertv/gophertunnel/master/gophertunnel_telescope_coloured.png)

## Overview

gophertunnel is composed of several packages that may be of use for creating Minecraft related tools.
As of version v1.38.0, Gophertunnel requires at least Go 1.22.
A brief overview of all packages may be found [here](https://pkg.go.dev/mod/github.com/sandertv/gophertunnel?tab=packages).

## Examples
Examples on how to dial a connection or start a server can be found in the [minecraft package](https://github.com/Sandertv/gophertunnel/tree/master/minecraft).
Additionally, a MITM proxy is implemented in the [main.go file](https://github.com/Sandertv/gophertunnel/blob/master/main.go).

## Versions
Gophertunnel supports only one version at a time (generally the latest official Minecraft release), but multiple protocols can be supported with the API. Generally, a new
minor version is tagged when gophertunnel supports a new Minecraft version that was not previously supported.

## Contact
[![Chat on Discord](https://img.shields.io/badge/Chat-On%20Discord-738BD7.svg?style=for-the-badge)](https://discord.com/invite/U4kFWHhTNR)

### Note: We do not under any circumstance support or endorse the usage of gophertunnel with malicious intent.
