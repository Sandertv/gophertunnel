// Package query implements the UT3 query protocol as described on
// http://wiki.unrealadmin.org/UT3_query_protocol. It is composed of a handshake, followed by data sent
// by the server that responds to a query sent by the client.
// This UT3 query protocol is used by most Minecraft related servers.
// Package query specialises on Minecraft Bedrock Edition related queries. Querying other types of game
// servers is not guaranteed to work.
//
// Where some server softwares (most common public ones) support this query protocol, others do not. A
// different kind of 'query', which is supported by all servers, may be performed using the go-raknet library.
package query
