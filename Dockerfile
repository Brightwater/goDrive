FROM golang:1.22

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go mod download

RUN go build -o goDrive main.go

CMD ["./goDrive"]