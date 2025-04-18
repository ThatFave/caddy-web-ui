FROM golang:1.24 AS builder

LABEL Maintainer="Michael Hesemann <michael@mhesemann.de>"

COPY src /src

WORKDIR /src

RUN go build .

FROM busybox

WORKDIR /

COPY --from=builder /src/caddy-web-ui .

COPY --from=builder /src/static /static

ENV CADDY_API=""

EXPOSE 8080

VOLUME [ "/config" ]

ENTRYPOINT [ "/caddy-web-ui" ]
