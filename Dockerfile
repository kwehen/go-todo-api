FROM golang:latest

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -v -o main .

CMD ["/app/main"]
# ENTRYPOINT [ "/app/main" ]