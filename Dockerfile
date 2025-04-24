FROM golang:1.24.0 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /soldatovadew .


FROM ubuntu:latest
WORKDIR /app
COPY --from=builder /soldatovadew /app/
COPY web ./web/
COPY .env .
EXPOSE 7540
ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db
VOLUME /data
CMD ["./soldatovadew"]



