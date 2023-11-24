const express = require("express");
const sqlite3 = require("sqlite3").verbose();

const app = express();
const request = require("request-promise");
const cors = require("cors");
const base_url = process.env.BASE_URL || "http://localhost";

app.use(cors());
app.use(express.json());

let port = process.env.PORT || "9090";
let dbPath;
switch (port) {
  case "9090":
    dbPath = 'kafka-events-0.db';
    break;
  case "9091":
    dbPath = 'kafka-events-1.db';
    break;
  case "9092":
    dbPath = 'kafka-events-2.db';
    break;
  default:
    console.error('Unsupported port:', port);
    process.exit(1);
}

// open database in memory
let db = new sqlite3.Database(dbPath, (err) => {
  if (err) {
    return console.error(err.message);
  }
  console.log(`Connected to the ${dbPath} SQLite database.`);
});

// Create 'events' table if it doesn't exist
let createTableSql = `
CREATE TABLE IF NOT EXISTS events (
  Minute INT NOT NULL,
  Player TEXT NOT NULL,
  Club TEXT NOT NULL,
  Score TEXT NOT NULL
);`;

db.run(createTableSql, (err) => {
  if (err) {
    console.error("Error creating table: ", err.message);
  } else {
    console.log("Events table created or already exists.");
  }
});

app.post("/addScore", (req, res) => {
  console.log("Adding Score", req.body);

  let assignedPort = getAssignedPort(req);
  const incomingScore = createIncomingScore(req, assignedPort);

  let z_ports = [8080, 8081, 8082];
  if (assignedPort) {
    z_ports = z_ports.filter(port => port !== assignedPort);
  }

  sendRequest(assignedPort || z_ports.shift(), incomingScore, z_ports, res)
      .then(response => {
        res.sendStatus(200);
      })
      .catch(error => {
        handleRequestError(assignedPort, incomingScore, z_ports, res);
      });
});

function getAssignedPort(req) {
  if (req.body && req.body.metadata && req.body.metadata.ReceiverIp) {
    return parseInt(req.body.metadata.ReceiverIp, 10);
  }
  return undefined;
}

function createIncomingScore(req, assignedPort) {
  const currentTimestamp = new Date().toISOString();
  return {
    Timestamp: currentTimestamp,
    Metadata: {
      SenderIp: port,
      ReceiverIp: assignedPort ? assignedPort.toString() : undefined,
      Timestamp: currentTimestamp,
      Version: 1,
      Attempts: 1,
      Clients: "9090,9091,9092"
    },
    GameResults: {
      Minute: req.body.gameResults.Minute,
      Player: req.body.gameResults.Player,
      Club: req.body.gameResults.Club,
      Score: req.body.gameResults.Score,
    },
  };
}

function sendRequest(currentPort, incomingScore, availablePorts, res) {
  incomingScore.Metadata.ReceiverIp = currentPort.toString();

  const options = {
    method: "POST",
    uri: base_url + ":" + currentPort + "/metadata",
    body: incomingScore,
    json: true,
    headers: {
      "Content-Type": "application/json",
    },
  };

  return new Promise((resolve, reject) => {
    request(options)
        .then(response => {
          resolve(response);
        })
        .catch(error => {
          reject(error);
        });
  });
}


function handleRequestError(failedPort, incomingScore, availablePorts, res) {
  console.log("Failed on port:", failedPort);
  availablePorts = availablePorts.filter(port => port !== failedPort);

  if (availablePorts.length > 0) {
    const nextPort = availablePorts.shift();
    console.log("Trying next port:", nextPort);

    sendRequest(nextPort, incomingScore, availablePorts, res)
        .then(response => {
          res.sendStatus(200);
        })
        .catch(error => {
          handleRequestError(nextPort, incomingScore, availablePorts, res);
        });
  } else {
    console.log("No more ports to try.");
    res.status(500).send("Internal Server Error");
  }
}

// Define a route to handle GET requests
app.get("/data", (req, res) => {
  db.all("SELECT * FROM events", (err, rows) => {
    if (err) {
      return console.error(err.message);
    }

    // This will send an array of rows to the frontend.
    res.send(rows);
  });
});

// Serve a simple HTML page with the array data
app.get("/", (req, res) => {
  db.all("SELECT * FROM events", (err, rows) => {
    if (err) {
      return console.error(err.message);
    }

    // This will log an array of rows to the console.
    console.log(rows);

    const htmlResponse = `
    <html>
      <body>
        <h1>Data:</h1>
        <pre>${JSON.stringify(rows, null, 2)}</pre>
      </body>
    </html>
  `;
    res.send(htmlResponse);
  });
});

app.post("/updateScore", (req, res) => {
  reqBody = req.body;
  console.log("Updating Score", req.body)
  try {
    db.run("INSERT INTO events (Minute, Player, Club, Score) VALUES (@minute, @player, @club, @score)", {
      "@minute": reqBody.Minute,
      "@player": reqBody.Player,
      "@club": reqBody.Club,
      "@score": reqBody.Score,
    });
  } catch {
    db.rollback();
  }

  res.status(200).send("Score Updated");
});

// Start the server
app.listen(port, () => {
  console.log(`Server is running on http://localhost:${port}`);
});