# Vesperis MultiProxy

A proxy network using Redis and Postgres to easily handle thousands of players shared accross multiple Gate proxies.

## Features
- Shared players. Each proxy has access to every player, even offline, to use efficiently.
- Ban system so an admin on proxy A can ban a player on proxy B with ease.