const express = require("express");
const sqlite3 = require("sqlite3").verbose();

const app = express();
const port = 9092;
const cors = require("cors");
const request = require("request-promise");

app.use(cors());
app.use(express.json());

// open database in memory
let db = new sqlite3.Database("kafka-events.db", (err) => {
  if (err) {
    return console.error(err.message);
  }
  console.log("Connected to the in-memory SQlite database.");
});

// const dataArray = [
//   { minutes: 28, player: "Leroy Sane", club: "FCB", score: "0-1" },
//   { minutes: 49, player: "Serge Gnabry", club: "FCB", score: "0-2" },
//   { minutes: 53, player: "Rasmus Hojlund", club: "MNU", score: "1-2" },
//   { minutes: 88, player: "Harry Kane", club: "FCB", score: "1-3" },
//   { minutes: 92, player: "Casemiro", club: "MNU", score: "2-4" },
//   { minutes: 95, player: "Mathys Tel", club: "FCB", score: "2-4" },
//   { minutes: 95, player: "Casemiro", club: "MNU", score: "3-4" },
// ];

app.post("/addScore", (req, res) => {
  const currentTimestamp = new Date().toISOString(); // Get the current time in RFC3339 format

  incomingScore = {
    Metadata: {
      SenderIp: req.body.metadata.SenderIp,
      ReceiverIp: req.body.metadata.ReceiverIp,
      Timestamp: currentTimestamp,
      Attempts: req.body.metadata.Attempts,
    },
    GameResults: {
      Min: req.body.gameResults.Min,
      Player: req.body.gameResults.Player,
      Club: req.body.gameResults.Club,
      Score: req.body.gameResults.Score,
    },
  };

  const options = {
    method: "POST",
    uri: "http://localhost:8080/score",
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
      console.log(error);
    }
  });
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
  try {
    db.run("INSERT INTO events (Min, Player, Club, Score) VALUES (@minute, @player, @club, @score)", {
      "@minute": reqBody.Min,
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
