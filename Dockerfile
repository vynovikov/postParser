FROM golang:1.20-buster as build

WORKDIR /postParser

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o postParser ./cmd/postParser

CMD ./postParser

FROM alpine:latest as release

RUN apk --no-cache add ca-certificates && \
	mkdir /tls


COPY --from=build /postParser ./ 

RUN chmod +x ./postParser

ENTRYPOINT [ "./postParser" ]

EXPOSE 443 3000