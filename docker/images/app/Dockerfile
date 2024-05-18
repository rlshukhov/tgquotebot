FROM debian:latest

RUN apt-get update && apt-get install -y \
    wget \
    ca-certificates \
    curl \
    webp \
    dnsutils

COPY ./build/app /app
RUN chmod +x /app
COPY ./build/quote.html /quote.html

RUN mkdir /tmp/quote
COPY ./build/font.ttf /tmp/quote/font.ttf
COPY ./build/avatar-placeholder.png /tmp/quote/avatar-placeholder.png

CMD ["/app"]