// Package query implements the UT3 query protocol as described on
// http://wiki.unrealadmin.org/UT3_query_protocol. It is composed of a handshake, followed by data sent
// by the server that responds to a query sent by the client.
//
// Where some server softwares (most common public ones, such as PocketMine) support this query protocol,
// others do not. A different kind of 'query', which is supported by all servers, may be performed using the
// go-raknet library. (raknet.Ping()) Server softwares which do not implement the query protocol include the
// Bedrock Dedicated Server.
package query
