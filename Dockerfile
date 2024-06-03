FROM ubuntu:latest

MAINTAINER wladbelsky <raf520wb@gmail.com>

ARG TARGETARCH

RUN mkdir "/app"

# texas_bot is the name of the binary file
# refer to github actions for the binary file name
COPY ./texas_bot_${TARGETARCH} /app/texas_bot

RUN chmod +x /app/texas_bot

WORKDIR /app

ENTRYPOINT ["/app/texas_bot"]
