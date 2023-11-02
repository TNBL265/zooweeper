 # ZooWeeper: 50.041 Distributed System Project

## Technical Diagrams
### High-Level Architecture
![](assets/system_architecture.png)

### ZooKeeper Internal Architecture
![](assets/zookeeper_internal_architecture.jpg)

### Implementation Focus (Checkpoint 2)
![](assets/request_processor_flow.png)

## ZooKeeper Internals
### 1. Data Synchronization: [./server/zab](./server/zab/zab.go)
#### Atomic Broadcast Protocol
- **Reliable delivery**:
  - using `WriteOpsMiddleware` to process all Write request:
    - All Write requests to Follower is forwarded to Leader
    - Leader only commit Write request once all Followers acknowledged
- **Total order**:
  - using state `ProposalState` and mutex `proposalMu`
    - requests are processed according to state changes
- **Causal order**:
  - using a min Priority Queue to order `RequestItem` by timestamp from Client's requests
  - assumption: no clock synchronization issue between Clients
#### Linearization Write and FIFO Client Order
- By ensuring 3 properties above
### 2. Distributed Coordination
### 3. Fault Tolerance

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
  - 3x Kafka Server (Express)
  - 2x Kafka Client Application (React)
- Should delete all your `.db` file else the build will be very slow
- Run:
```shell
cd zooweeper
docker-compose up
```
- Test sending of `100` requests
```shell
cd zooweeper
./send_requests.sh 100
```

## References:
- [Zookeeper Internals](https://zookeeper.apache.org/doc/r3.9.0/zookeeperInternals.html)
- [Apache Zookeeper Java implementation](https://github.com/apache/zookeeper)
- [Zookeeper Paper](https://pdos.csail.mit.edu/6.824/papers/zookeeper.pdf)
- [Zab Paper](https://ieeexplore.ieee.org/stamp/stamp.jsp?arnumber=5958223)
- [Native Go Zookeeper Client Library](https://github.com/go-zookeeper/zk)
