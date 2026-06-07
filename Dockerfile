FROM golang:1.22-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/nv-copa ./cmd/server

FROM alpine:3.20

WORKDIR /app
RUN apk add --no-cache ca-certificates sqlite tzdata
COPY --from=build /out/nv-copa /app/nv-copa
COPY web /app/web
COPY scripts /app/scripts

ENV ADDR=:8080
ENV DATABASE_PATH=/data/copa.db

EXPOSE 8080
VOLUME ["/data", "/backups"]

CMD ["/app/nv-copa"]
