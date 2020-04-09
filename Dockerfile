FROM golang:1.14 AS builder
WORKDIR /source
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o mercanabo .

FROM alpine:latest
RUN apk --no-cache add tzdata ca-certificates postgresql-client
WORKDIR /bot/
COPY --from=builder /source/texts texts
COPY --from=builder /source/mercanabo .
CMD ["./mercanabo"]
