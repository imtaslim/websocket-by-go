FROM alpine:3.13

WORKDIR /root/

COPY websocket websocket
ADD env/sample.config env/config

ENTRYPOINT ["./websocket"]