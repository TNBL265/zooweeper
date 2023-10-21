// src/DataDisplay.js
import React, { useState, useEffect } from 'react';
import axios from 'axios';

function DataDisplay() {
  const [data, setData] = useState([]);
  const port = 9092;

  useEffect(() => {
    // HTTP GET request to your Kafka server's /data endpoint
    axios.get(`http://localhost:${port}/data`)
      .then((response) => {
        setData(response.data);
      })
      .catch((error) => {
        console.error('Error fetching data:', error);
      });
  }, []);

  return (
    <div>
      <h1>Data Display from Kafka Server</h1>
      <ul>
        {data.map((item, index) => (
          <li key={index}>
            {item.minutes} mins - {item.player} ({item.club}) - Score: {item.score}
          </li>
        ))}
      </ul>
    </div>
  );
}

export default DataDisplay;
