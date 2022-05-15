FROM golang:1.18.0-buster as builder

RUN apt-get update && \
    apt-get install -y \
    ffmpeg \
    libopus-dev

COPY ./ /app
WORKDIR /app
RUN make build
CMD /app/ceviord
