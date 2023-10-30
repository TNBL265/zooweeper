PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;

CREATE TABLE ZNode (
  NodeId INTEGER PRIMARY KEY AUTOINCREMENT,
  LeaderServer TEXT,
  Servers TEXT,
  NodeIp TEXT,
  SenderIp TEXT,
  ReceiverIp TEXT,
  Timestamp DATETIME,
  Attempts INTEGER,
  Version INTEGER,
  ParentId INTEGER
);

INSERT INTO ZNode (LeaderServer, Servers, NodeIp, SenderIp, ReceiverIp, Timestamp, Attempts, Version, ParentId) VALUES ('9090', '9090, 9091, 9092','8080', '-', '-', CURRENT_TIMESTAMP, 1, 0, 0);
INSERT INTO ZNode (LeaderServer, Servers, NodeIp, SenderIp, ReceiverIp, Timestamp, Attempts, Version, ParentId) VALUES ('-', '-', '8080', 'postman', '9090', CURRENT_TIMESTAMP, 2, 0, 2);
INSERT INTO ZNode (LeaderServer, Servers, NodeIp, SenderIp, ReceiverIp, Timestamp, Attempts, Version, ParentId) VALUES ('-', '-', '8080', 'postman', '9090', CURRENT_TIMESTAMP, 3, 0, 3);

COMMIT;
