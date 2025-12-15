# Vesperis MultiProxy
A proxy network using Redis and Postgres to easily handle thousands of players shared across multiple Gate proxies.

## Features
- Ban System.
Ban players from every proxy, with the ability to optionaly ban players temporarily for a limited time.

- Friend System.
Players can invite others to become friends. Friends can see information like last seen.

- Party System.
Players can create parties and invite others.

- Vanish Stystem.
Players can become hidden on the server by becoming vanished.

## How does it work?
-  Shared playerdata. 
Each proxy has access to every player, even offline, to use efficiently. 

- Communicating proxies.
If one proxy goes offline, players that are located on that proxy will be placed on other proxies automatically. Send tasks from proxy A to proxy B and get feedback if it was successful.

- Efficient cache. 
Information is stored per proxy and automatically changed when needed. This makes it use the database less and making the proxy as fast as possible.
