FROM golang

WORKDIR /app

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /my_app

ENV TODO_PASSWORD="1234"
ENV TODO_DBFILE="sheduler"
ENV TODO_PORT="7540"

EXPOSE $TODO_PORT:$TODO_PORT

CMD ["/my_app"]