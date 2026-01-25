FROM golang:1.25.3 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/service ./cmd/service && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/worker ./cmd/worker

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /app/bin/service /service-courier
COPY --from=builder /app/bin/worker /worker-courier
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/service-courier"]
