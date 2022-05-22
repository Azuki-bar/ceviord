FROM golang:1.18.2-bullseye as builder

RUN apt-get update  \
 && \
    apt-get install -y --no-install-recommends \
    ffmpeg \
    libopus-dev \
  &&  \
    apt-get clean \
  && \
    rm -rf /var/lib/apt/lists/*

COPY ./ /app
WORKDIR /app
RUN make build
CMD /app/ceviord
