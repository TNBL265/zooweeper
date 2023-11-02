FROM golang:1.19.2-bullseye

WORKDIR /app
COPY go.mod .
RUN go mod tidy
COPY . .

CMD ["go", "run", "main.go"]