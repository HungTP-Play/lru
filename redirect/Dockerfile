FROM golang:1.19-bullseye

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o redirect .

CMD ["./redirect"]