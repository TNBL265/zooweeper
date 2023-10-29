# ZooWeeper: 50.041 Distributed System Project

## Technical Diagrams
### High-Level Architecture
![image](https://github.com/TNBL265/zooweeper/assets/84057800/a1ca45c3-08ba-4fa9-ae25-aa616a9d0f40)
### Implementation Focus
![image](https://github.com/TNBL265/zooweeper/assets/84057800/f4c8fc50-a734-4d08-89b4-28e20041991d)

## Running Locally
Each of the 3 applications have a Dockerfile on its root, running docker compose up would build the image for each application. 

- 3x Zookeeper Server
- 2x Kafka Server (Express)
- 1x Kafka React Application

1. Create SQLite3 db: `sqlite3 kafka-db.sqlite < kafka-db.sql`
2. Switch to the root directory and run `docker-compose up -d`

## Structure Overview
- `client` - Implement client API and connection
- `server` - Implement main server features:
  - `core`: server connection, session management, watcher and request processors flow for leader and followers
  - `database`: main data structure
  - `request_processors`: ensure linearizable writes and FIFO client order
  - `zab`: fault tolerance, leader election
```bash
zooweeper/
│
├── client/
│   ├── client_connection.go
│   ├── zw.go
│
├── server/
│   ├── core/
│   │   ├── follower_zw_server.go
│   │   ├── leader_zw_server.go
│   │   ├── server_connection.go
│   │   ├── session_tracker.go
│   │   ├── watcher.go
│   │   └── zw_server.go
│   ├── database/
│   │   ├── znode.go
│   │   ├── ztree.go
│   ├── request_processors/
│   │   ├── common_processors.go
│   │   ├── follower_processors.go
│   │   ├── leader_processors.go
│   │   ├── request_processor.go
│   ├── zab/
│   │   ├── follower.go
│   │   ├── follower_handler.go
│   │   ├── leader.go
│   │   ├── leader_election.go
│   │   ├── quorum.go
```

## References:
- [Apache Zookeeper Java implementation](https://github.com/apache/zookeeper)
- [Zookeeper Paper](https://pdos.csail.mit.edu/6.824/papers/zookeeper.pdf)
- [Zab Paper](https://ieeexplore.ieee.org/stamp/stamp.jsp?arnumber=5958223)
- [Native Go Zookeeper Client Library](https://github.com/go-zookeeper/zk)
