FROM golang:1.17.0-alpine3.14 as build
LABEL org.opencontainers.image.source="https://github.com/Nikki18977/0101_scratch"

ENV USER=appuser
ENV UID=10001

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/bin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

RUN mkdir /uploads

WORKDIR  /go/src/app
COPY  app/ .

RUN go mod download && go mod verify
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o /go/bin/app.bin cmd/main.go

FROM busybox:1.35.0-uclibc as busybox
FROM scratch

COPY --from=busybox /bin/sh /bin/sh
COPY --from=busybox /bin/chown /bin/chown
COPY --from=build /uploads /uploads
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /go/bin/app.bin /go/bin/app.bin

RUN chown -R appuser:appuser /uploads 
    
VOLUME [ "/uploads" ]

USER appuser:appuser

EXPOSE 9999

ENTRYPOINT ["/go/bin/app.bin"]
