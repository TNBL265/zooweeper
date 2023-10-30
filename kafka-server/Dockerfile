FROM node:18

WORKDIR /kafka

COPY package.json .
COPY /kafka-events.db .

RUN npm install

COPY . .

EXPOSE 9090

CMD ["npm", "start"]