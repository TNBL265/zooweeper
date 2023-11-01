 # ZooWeeper: 50.041 Distributed System Project

## Technical Diagrams
### High-Level Architecture
![image](https://github.com/TNBL265/zooweeper/assets/84057800/67a7d3ea-9586-4ad3-8701-db11fa0294db)

### Implementation Focus (Checkpoint 1)
![image](https://github.com/TNBL265/zooweeper/assets/84057800/a27402fe-e84e-414c-a61b-307a840be2f2)


## Local development
### Postman
- Start a new Postman workspace: https://web.postman.co/workspaces
- Import **Collections** from [./postman](./postman)
- Set **Environments**:
  - `goBaseUrl=8080`
  - `expressBaseUrl=9090`
### Zookeeper Server
- Run: 
```shell
cd zooweeper/server
go mod tidy 
PORT=8080 go run main.go
```
- Output: `pong` on `localhost:8080`
### Kafka Server (Express)
- Create database:
```shell
cd zooweeper/kafka-server
sqlite3 kafka-events-0.db < kafka-events.sql
```
- Run:
```shell
cd zooweeper/kafka-server
npm install
PORT=9090 npm start
```
- Output: `Events` json on `localhost:9090`
### Kafka Client Application (React)
- Run: 
```shell
cd zooweeper/kafka-react-app
npm install
PORT=3000 npm start
```
- Output: formatted `Events` json on `localhost:3000`  (when Kafka Server is running)

### Distributed System Demo
- Overview: The above applications would be dockerized:
  - 3x Zookeeper Server
  - 2x Kafka Server (Express)
  - 1x Kafka Client Application (React)
- Run:
```shell
cd zooweeper
docker-compose up
```

## References:
- [Apache Zookeeper Java implementation](https://github.com/apache/zookeeper)
- [Zookeeper Paper](https://pdos.csail.mit.edu/6.824/papers/zookeeper.pdf)
- [Zab Paper](https://ieeexplore.ieee.org/stamp/stamp.jsp?arnumber=5958223)
- [Native Go Zookeeper Client Library](https://github.com/go-zookeeper/zk)
