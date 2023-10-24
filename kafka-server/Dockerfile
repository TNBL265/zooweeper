FROM node:18

WORKDIR /kafka

COPY package.json .
COPY /kafka-db.sqlite .

RUN npm install

COPY . .

EXPOSE 9092

CMD ["node", "server.js"]