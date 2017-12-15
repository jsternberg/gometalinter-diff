FROM golang:1.9.2 as builder
RUN go get -u github.com/alecthomas/gometalinter && \
    gometalinter --install
COPY . /go/src/github.com/jsternberg/gometalinter-diff
RUN go get github.com/jsternberg/gometalinter-diff && \
    go install github.com/jsternberg/gometalinter-diff
ENTRYPOINT ["gometalinter-diff"]
