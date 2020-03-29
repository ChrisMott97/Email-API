FROM golang:latest

WORKDIR /app

COPY . /app

RUN go mod download

RUN go get github.com/githubnemo/CompileDaemon

ENTRYPOINT /go/bin/CompileDaemon --build="go build main.go" --command=./main