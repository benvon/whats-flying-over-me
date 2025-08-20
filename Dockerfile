FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY whats-flying-over-me /usr/local/bin/whats-flying-over-me
ENTRYPOINT ["/usr/local/bin/whats-flying-over-me"]
