FROM golang:1.19.0-bullseye

COPY src/ /root/ethereum-account-generator/

WORKDIR /root/ethereum-account-generator/

RUN go get ./... && make build

CMD ["/root/ethereum-account-generator/bin/generator"]
