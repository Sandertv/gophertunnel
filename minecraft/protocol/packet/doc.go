// Package packet implements every packet in the Minecraft Bedrock Edition protocol using the functions found
// in the minecraft/protocol package. Each of the packets hold their own encoding and decoding methods which
// are used to read the packet from a bytes.Buffer or write the packet to one.
//
// Besides the implementations of packets themselves, the packet package also implements the decoding and
// encoding of the lowest level Minecraft related packets, meaning the compressed packet batches. It handles
// the compression and (optional) encryption of these packet batches.
package packet
