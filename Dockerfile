FROM golang:1.13.3-alpine3.10
ENV GOPATH="/go"
RUN ["mkdir", "-p", "/go/src/github.com/danielorf/henrylibrary"]
COPY * /go/src/github.com/danielorf/henrylibrary/
WORKDIR /go/src/github.com/danielorf/henrylibrary
RUN ["apk", "add", "git"]
RUN ["go", "build", "-o", "henrylibrary"]
CMD ["./henrylibrary"]
