# Vesperis MultiProxy

A proxy network using Redis and Postgres to easily handle thousands of players shared across multiple Gate proxies.

## Features
- Shared players. Each proxy has access to every player, even offline, to use efficiently. Example: ban a player that plays on proxy A from proxy B with ease.
- Communicating proxies. If one proxy goes offline, players that are located on that proxy will be placed on other proxies automatically.
- Efficient cache. Use the database as little as possible, making the proxy network as fast as possible.