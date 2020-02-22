FROM golang:latest
LABEL maintainer="Daniel Orf <danielorf@gmail.com>"
WORKDIR /app

RUN useradd -ms /bin/bash docker

COPY go.mod go.sum ./
RUN go mod download
COPY --chown=docker:docker . .
RUN go build -o henrylibrary .
EXPOSE 3000

USER docker
CMD ["./henrylibrary"]
