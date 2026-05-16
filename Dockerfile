FROM golang:1.26.1-alpine AS builder 

WORKDIR /app 

COPY go.mod go.sum ./ 
RUN go mod download 

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /polls-api ./cmd/api/main.go 
RUN CGO_ENABLED=0 GOOS=linux go build -o /polls-reporter ./cmd/reporter/main.go

FROM alpine:latest 

RUN apk add --no-cache bash

WORKDIR /

COPY --from=builder /polls-api /polls-api 
COPY --from=builder /polls-reporter /polls-reporter 
COPY scripts/init.sql /init.sql

RUN echo '#!/bin/bash' > /start.sh && \
    echo '/polls-reporter &' >> /start.sh && \
    echo 'exec /polls-api' >> /start.sh && \
    chmod +x /start.sh

EXPOSE 8080
EXPOSE 50051 
CMD ["/start.sh"]