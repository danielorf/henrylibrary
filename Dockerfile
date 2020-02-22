FROM golang:latest
LABEL maintainer="Daniel Orf <danielorf@gmail.com>"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o henrylibrary .
EXPOSE 3000
CMD ["./henrylibrary"]
