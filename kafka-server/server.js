const express = require("express");
const sqlite3 = require("sqlite3").verbose();

const app = express();
const request = require("request-promise");
const cors = require("cors");

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
  console.log("Adding Score", req.body)
  const currentTimestamp = new Date().toISOString(); // Get the current time in RFC3339 format

  incomingScore = {
    Timestamp: currentTimestamp,
    Metadata: {
      SenderIp: port,
      ReceiverIp: req.body.metadata.ReceiverIp.toString(),
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

  let assignedPort = parseInt(req.body.metadata.ReceiverIp, 10);
  let z_ports = [8080,8081,8082]
  const base_url = "http://localhost"

  function sendRequest(assignedPort) {

    const options = {
      method: "POST",
      uri: base_url + ":" + assignedPort + "/metadata",
      body: incomingScore,
      json: true,
      headers: {
        "Content-Type": "application/json",
      },
    };

    request(options, function (error, response, body) {
      if (!error) {
        res.sendStatus(200);
      } else {
        // console.log(error);
        if (assignedPort !== undefined) {
          console.log("next port:", assignedPort);
          z_ports = z_ports.filter(port => port !== assignedPort);
          console.log(z_ports);
          assignedPort = z_ports[0]
          sendRequest(assignedPort);
        } else {
          console.log("No more ports to try.");
          res.sendStatus(404)
        }
      }
    });

  }
  z_ports = z_ports.filter(port => port !== assignedPort);
  sendRequest(assignedPort);
});

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