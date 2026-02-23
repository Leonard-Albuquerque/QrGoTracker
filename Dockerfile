FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /qr-tracker ./cmd/server
EXPOSE 8085
CMD ["/qr-tracker"]
