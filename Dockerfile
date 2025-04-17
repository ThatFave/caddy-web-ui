FROM golang:1.24 AS builder

LABEL Maintainer="Michael Hesemann <michael@mhesemann.de>"

COPY src /src

WORKDIR /src

RUN go build .

FROM busybox

COPY --from=builder /src/caddy-web-ui .

ENV CADDY_API=""

EXPOSE 8080

VOLUME [ "/config" ]

ENTRYPOINT [ "caddy-web-ui" ]
