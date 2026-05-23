// Package p2p provides client-side helpers for discovering and joining
// player-hosted Minecraft worlds advertised through Xbox Live multiplayer
// sessions.
//
// A Client lists joinable worlds from MPSD activity handles. Joining a world
// enters the advertised multiplayer session and waits for the host to publish a
// nonce for the joining player's XUID. Callers must copy Session.Nonce into
// login.ClientData.Nonce when dialing the host; vanilla hosts use this nonce to
// bind the Minecraft login request to the MPSD join.
//
// The package intentionally focuses on joining existing worlds. Publishing or
// hosting worlds requires keeping the advertised session status, connection
// details, player counts, and per-player nonces in sync with the running
// listener and is not implemented here.
package p2p
