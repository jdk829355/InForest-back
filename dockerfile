FROM golang:1.25-alpine3.21

RUN mkdir /app

WORKDIR /app

COPY ./bin/server_amd64 /app/server

CMD ["/app/server"] 