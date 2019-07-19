# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from golang v1.11 base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="Rajeev Singh <rajeevhub@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/callicoder/go-docker

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Download all the dependencies
# https://stackoverflow.com/questions/28031603/what-do-three-dots-mean-in-go-command-line-invocations
RUN go get -d -v ./...

# Install the package
RUN go build -o main .

RUN adduser -S -D -H -h /app appuser
USER appuser

# This container exposes port 8080 to the outside world
EXPOSE 8080

CMD ["./main"]
