FROM golang:1.26.1-alpine AS builed 

WORKDIR /app 

COPY go.mod go.sum ./ 
RUN go mod download 

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /polls-api ./cmd/api/main.go 

FROM alpine:latest 
WORKDIR /
COPY --from=builder /polls-api /polls-api 

EXPOSE 8080 
CMD ["/polls-api"]