PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;

CREATE TABLE znode (
  NodeId INTEGER PRIMARY KEY AUTOINCREMENT,
  LeaderServer TEXT,
  Servers TEXT,
  SenderIp TEXT,
  ReceiverIp TEXT,
  Timestamp DATETIME,
  Attempts INTEGER,
  ParentId INTEGER
);


INSERT INTO znode (LeaderServer, Servers, SenderIp, ReceiverIp, Timestamp, Attempts, ParentId) VALUES ('192.168.1.1', '192.168.1.2, 192.168.1.3', '192.168.1.4', '192.168.1.5', CURRENT_TIMESTAMP, 1, 0);
INSERT INTO znode (LeaderServer, Servers, SenderIp, ReceiverIp, Timestamp, Attempts, ParentId) VALUES ('192.168.1.1', '192.168.1.2, 192.168.1.3', '192.168.1.8', '192.168.1.9', CURRENT_TIMESTAMP, 2, 2);
INSERT INTO znode (LeaderServer, Servers, SenderIp, ReceiverIp, Timestamp, Attempts, ParentId) VALUES ('192.168.1.1', '192.168.1.2, 192.168.1.3', '192.168.1.10', '192.168.1.11', CURRENT_TIMESTAMP, 3, 3);

COMMIT;