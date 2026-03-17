FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/cos-mock ./tests/mock-server/cos

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /out/cos-mock /app/cos-mock
EXPOSE 9000
ENTRYPOINT ["/app/cos-mock"]

