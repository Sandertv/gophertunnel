// Package minecraft implements Minecraft Bedrock Edition connections. It implements a Dial() function to dial
// a connection to a Minecraft server and a Listen() function to create a listener in order to listen for
// incoming Minecraft connections. Typically these connections are done over RakNet, which is implemented by
// the github.com/sandertv/go-raknet package.
//
// The minecraft package provides a high level abstraction over Minecraft network related features. It handles
// the authentication, encryption and spawning sequence by itself and users can send and read packets
// implemented in the minecraft/protocol/packet package.
package minecraft
