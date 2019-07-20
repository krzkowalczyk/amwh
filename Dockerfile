FROM golang:1.12.7-alpine as builder

WORKDIR /go/src/app
COPY . .
RUN apk add git 
RUN go get -d -v ./...
RUN go build -o main .

FROM alpine

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /go/src/app/main /app/
WORKDIR /app
EXPOSE 8080
CMD ["./main"]