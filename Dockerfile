# ssl certs for sending webhooks to discord.com
FROM alpine as certs
RUN apk update && apk add ca-certificates

# builder image
FROM golang:1.20-alpine as builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o discoroll .

# generate clean, final image for end users
FROM busybox
COPY --from=certs /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /build/discoroll .

EXPOSE 8080

# executable
CMD [ "./discoroll" ]
