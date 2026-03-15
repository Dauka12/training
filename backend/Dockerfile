FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

FROM alpine:3.21
RUN adduser -D appuser
WORKDIR /app
COPY --from=build /out/api /app/api
COPY migrations /app/migrations
USER appuser
EXPOSE 8080
CMD ["/app/api"]
