FROM golang:1.22.0-alpine

WORKDIR /app

COPY . .

RUN apk add unzip
RUN mkdir /config

RUN go mod download
RUN go build

RUN wget -O bw_cli.zip https://github.com/bitwarden/clients/releases/download/cli-v2024.4.1/bw-linux-2024.4.1.zip
RUN unzip bw_cli.zip -d /usr/local/bin/

ENTRYPOINT ["/app/bitwarden-backup"]