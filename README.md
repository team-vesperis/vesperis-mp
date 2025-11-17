# Vesperis MultiProxy
A proxy network using Redis and Postgres to easily handle thousands of players shared across multiple Gate proxies.

## Features
-  Shared playerdata. 

Each proxy has access to every player, even offline, to use efficiently. 

- Communicating proxies.

If one proxy goes offline, players that are located on that proxy will be placed on other proxies automatically. Send tasks from proxy A to proxy B and get feedback if it was successful.

- Efficient cache. 

Information is stored per proxy and automatically changed when needed. This makes it use the database less and making the proxy as fast as possible.