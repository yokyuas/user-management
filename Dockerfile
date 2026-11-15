FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/api ./cmd/api
FROM alpine:3.20
RUN adduser -D app
USER app
COPY --from=build /app/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
