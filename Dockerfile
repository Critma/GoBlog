FROM golang:1.25.1-alpine3.22 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api cmd/api/*.go

FROM scratch
WORKDIR /app
COPY --from=builder ./app/api .
COPY ./.env .
EXPOSE 8080
CMD ["./api"]