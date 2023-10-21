const express = require('express');

const app = express();
const port = 9092;
const cors = require('cors');

app.use(cors());

const dataArray = [
  {minutes:28,player:"Leroy Sane",club:"FCB",score:"0-1"},
  {minutes:49,player:"Serge Gnabry",club:"FCB",score:"0-2"},
  {minutes:53,player:"Rasmus Hojlund",club:"MNU",score:"1-2"},
  {minutes:88,player:"Harry Kane",club:"FCB",score:"1-3"},
  {minutes:92,player:"Casemiro",club:"MNU",score:"2-4"},
  {minutes:95,player:"Mathys Tel",club:"FCB",score:"2-4"},
  {minutes:95,player:"Casemiro",club:"MNU",score:"3-4"}
];

// Define a route to handle GET requests
app.get('/data', (req, res) => {
  res.json(dataArray); // Respond with the array as JSON
});

// Serve a simple HTML page with the array data
app.get('/', (req, res) => {
  const htmlResponse = `
    <html>
      <body>
        <h1>Data:</h1>
        <pre>${JSON.stringify(dataArray, null, 2)}</pre>
      </body>
    </html>
  `;
  res.send(htmlResponse);
});

// Start the server
app.listen(port, () => {
  console.log(`Server is running on http://localhost:${port}`);
});
