FROM ubuntu:20.04

MAINTAINER wladbelsky <raf520wb@gmail.com>

# texas_bot is the name of the binary file
# refer to github actions for the binary file name
ADD ./texas_bot /app/

WORKDIR /app

ENTRYPOINT ["./texas_bot"]
