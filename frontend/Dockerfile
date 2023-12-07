FROM node:18

WORKDIR /app
COPY public/ public
COPY src/ src
COPY package.json .
RUN npm install
COPY . .

CMD ["npm", "start"]
