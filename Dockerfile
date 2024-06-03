FROM alpine:latest

MAINTAINER wladbelsky <raf520wb@gmail.com>

RUN mkdir "/app"

# texas_bot is the name of the binary file
# refer to github actions for the binary file name
COPY ./texas_bot /app/texas_bot

WORKDIR /app

CMD "/app/texas_bot"
