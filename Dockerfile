FROM golang:1.17 AS builder
WORKDIR /source
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o mercanabo .

FROM alpine:latest
RUN apk --no-cache add tzdata ca-certificates postgresql-client
WORKDIR /bot/
COPY --from=builder /source/texts texts
COPY --from=builder /source/mercanabo .
CMD ["./mercanabo"]
