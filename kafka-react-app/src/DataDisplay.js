import React, { useState, useEffect } from "react";
import axios from "axios";

function DataDisplay() {
    const [data, setData] = useState([]);

    // Helper function to get a random port
    const getRandomPort = () => {
        const ports = [9090, 9091, 9092];
        return ports[Math.floor(Math.random() * ports.length)];
    };

    const port = getRandomPort();
    console.log("Kafka Server: ", port);

    useEffect(() => {
        // HTTP GET request to your Kafka server's /data endpoint
        axios
            .get(`http://localhost:${port}/data`)
            .then((response) => {
                setData(response.data);
            })
            .catch((error) => {
                console.error("Error fetching data:", error);
            });
    }, []);

    return (
        <div>
            <h1>Data Display from Kafka Server</h1>
            <ul>
                {data.map((item, index) => (
                    <li key={index}>
                        {item.Min} mins - {item.Player} ({item.Club}) - Score: {item.Score}
                    </li>
                ))}
            </ul>
        </div>
    );
}

export default DataDisplay;
