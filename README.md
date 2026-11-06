# Key-Value-Server

A key/value server for the MIT 6.824 lab. Clients read and update values over RPC, and use the server to coordinate locks even when the network drops messages.

Some key features that I implemented:

- In-memory storage with version checks to prevent stale writes
- Client Get and Put calls over RPC
- Lock acquire and release using conditional writes
- RPC retries and handling for writes that may have succeeded despite a lost reply
- Lock ownership checks to recover from dropped replies

The key/value and lock tests pass with Go's race detector enabled.
