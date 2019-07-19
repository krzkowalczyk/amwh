FROM golang:alpine
RUN mkdir /app 
ADD . /app/
WORKDIR /app 
RUN go get ./...; go build -o main .
RUN adduser -S -D -H -h /app appuser
USER appuser
CMD ["./main"]
