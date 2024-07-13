FROM golang:1.21.0

WORKDIR /Docker

COPY . .

RUN go build -o app

LABEL institute="Reboot01"

LABEL version="1.0"

LABEL description="forum"

EXPOSE 8800

CMD ["./app"]