import React, { useState, useEffect } from "react";
import axios from "axios";

const base_url = process.env.BASE_URL || "http://localhost";

function DataDisplay() {
    const [data, setData] = useState([]);

    // Helper function to get a random port
    const getRandomPort = () => {
        const ports = [9090, 9091, 9092];
        return ports[Math.floor(Math.random() * ports.length)];
    };

    const port = getRandomPort();
    console.log("Kafka Broker: ", port);

    useEffect(() => {
        // HTTP GET request to your Kafka broker's /data endpoint
        axios
            .get(`${base_url}:${port}/data`)
            .then((response) => {
                setData(response.data);
            })
            .catch((error) => {
                console.error("Error fetching data:", error);
            });
    }, []);

    return (
        <div>
            <h1>Ordered List of Football Goals Timeline</h1>
            <ul>
                {data.map((item, index) => (
                    <li key={index}>
                        {item.Minute} mins - {item.Player} ({item.Club}) - Score: {item.Score}
                    </li>
                ))}
            </ul>
        </div>
    );
}

export default DataDisplay;
