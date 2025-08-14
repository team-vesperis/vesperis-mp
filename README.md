# Vesperis MultiProxy

A proxy network using Redis and Postgres to easily handle thousands of players shared accross multiple Gate proxies.

## Features
- Shared players. Each proxy has access to every player, even offline, to use efficiently. Ban a player that plays on proxy A from proxy B with ease.
- Communicating proxies. Use caches to use less connections with the database and update the caches when needed. If one proxy goes offline, players that are located on that proxy will be placed on other proxies automatically.