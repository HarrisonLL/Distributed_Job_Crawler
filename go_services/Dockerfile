FROM golang:1.21-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk add --no-cache docker-cli
COPY --from=builder /app/main .
COPY ./static ./static
COPY ./templates ./templates
EXPOSE 8080
CMD ["./main"]
