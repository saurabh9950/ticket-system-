
FROM golang:1.22-alpine AS builder

WORKDIR /app


COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o ticket-system .

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/ticket-system .

EXPOSE 8080

CMD ["./ticket-system"]
