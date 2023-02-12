FROM golang:1.19.5-alpine3.17 as builder

RUN apk add --no-cache make

COPY . /app

WORKDIR /app

RUN make build

# Use the alpine image as the final image
FROM alpine

# Copy binaries
COPY --from=builder /app/bin/ /bin/

WORKDIR /bin/

CMD ["/bin/sh"]
