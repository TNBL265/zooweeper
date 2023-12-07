FROM node:18

WORKDIR /kafka
COPY package.json .
RUN npm install
COPY . .

CMD ["npm", "start"]