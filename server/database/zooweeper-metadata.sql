PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;

CREATE TABLE znode (
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


INSERT INTO znode (LeaderServer, Servers, NodeIp, SenderIp, ReceiverIp, Timestamp, Attempts, Version, ParentId) VALUES ('9092', '9092 9093, 9094','8080', '-', '-', CURRENT_TIMESTAMP, 1, 0, 0);
INSERT INTO znode (LeaderServer, Servers, NodeIp, SenderIp, ReceiverIp, Timestamp, Attempts, Version, ParentId) VALUES ('-', '-', '8080', 'postman', '9092', CURRENT_TIMESTAMP, 2, 0, 2);
INSERT INTO znode (LeaderServer, Servers, NodeIp, SenderIp, ReceiverIp, Timestamp, Attempts, Version, ParentId) VALUES ('-', '-', '8080', 'postman', '9092', CURRENT_TIMESTAMP, 3, 0, 3);

COMMIT;


-- BEGIN TRANSACTION;

-- CREATE TABLE znode (
--   NodeId INTEGER PRIMARY KEY AUTOINCREMENT,
--   LeaderServer TEXT,
--   SenderIp TEXT,
--   ReceiverIp TEXT,
--   Timestamp DATETIME,
--   Attempts INTEGER,
--   Version INTEGER,
--   ParentId INTEGER
-- );

-- CREATE TABLE ServerArray (
--   ServerId INTEGER PRIMARY KEY AUTOINCREMENT,
--   NodeId INTEGER,
--   Server TEXT,
--   FOREIGN KEY (NodeId) REFERENCES znode (NodeId)
-- );

-- -- Insert sample data into znode table
-- INSERT INTO znode (LeaderServer, SenderIp, ReceiverIp, Timestamp, Attempts, Version, ParentId) VALUES ('192.168.1.1', '192.168.1.4', '192.168.1.5', CURRENT_TIMESTAMP, 1, 0, 0);
-- INSERT INTO znode (LeaderServer, SenderIp, ReceiverIp, Timestamp, Attempts, Version, ParentId) VALUES ('192.168.1.1', '192.168.1.8', '192.168.1.9', CURRENT_TIMESTAMP, 2, 0, 2);
-- INSERT INTO znode (LeaderServer, SenderIp, ReceiverIp, Timestamp, Attempts, Version, ParentId) VALUES ('192.168.1.1', '192.168.1.10', '192.168.1.11', CURRENT_TIMESTAMP, 3, 0, 3);

-- -- Insert sample server data into ServerArray table
-- INSERT INTO ServerArray (NodeId, Server) VALUES (1, '192.168.1.2');
-- INSERT INTO ServerArray (NodeId, Server) VALUES (1, '192.168.1.3');
-- INSERT INTO ServerArray (NodeId, Server) VALUES (2, '192.168.1.2');
-- INSERT INTO ServerArray (NodeId, Server) VALUES (2, '192.168.1.3');
-- INSERT INTO ServerArray (NodeId, Server) VALUES (3, '192.168.1.2');
-- INSERT INTO ServerArray (NodeId, Server) VALUES (3, '192.168.1.3');

-- COMMIT;