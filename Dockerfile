FROM golang:1.18-alpine

WORKDIR /postParser

COPY . .

EXPOSE 3000

CMD go run cmd/postParser/postParser.go