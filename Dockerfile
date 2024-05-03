FROM golang:1.21 AS build_stage

WORKDIR /app
COPY go.mod go.sum ./
COPY pb/*.go ./pb/
COPY idjson/*.go ./idjson/
COPY *.go ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o mkv-utils-server .

FROM debian:bookworm-slim

COPY --chmod=0700 install_mkvtoolnix.sh ./
RUN ./install_mkvtoolnix.sh && rm ./install_mkvtoolnix.sh
COPY --from=build_stage /app/mkv-utils-server /mkv-utils-server
EXPOSE 25002
ENTRYPOINT ["/mkv-utils-server"]

