FROM golang:1.17.0-alpine3.14 as build

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

WORKDIR  /go/src/app

COPY app/ .
RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/app.bin cmd/main.go

FROM scratch

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /go/bin/app.bin /go/bin/app.bin

VOLUME [ "/upload" ]

USER appuser:appuser

EXPOSE 9999

ENTRYPOINT ["/go/bin/app.bin"]
