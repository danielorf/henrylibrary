FROM golang:1.12.1-alpine3.9
ENV GOPATH="/go"
RUN ["mkdir", "-p", "/go/src/github.com/danielorf/henrylibrary"]
COPY * /go/src/github.com/danielorf/henrylibrary/
WORKDIR /go/src/github.com/danielorf/henrylibrary
RUN ["go", "build", "-o", "henrylibrary"]
CMD ["./henrylibrary"]
