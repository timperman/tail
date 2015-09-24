FROM golang:1.5.1

EXPOSE 8080
COPY . /go/src/github.com/timperman/tail

RUN go get -d -v ./...

RUN go test github.com/timperman/tail && go install ./...

ENTRYPOINT [ "/go/bin/tail" ]
