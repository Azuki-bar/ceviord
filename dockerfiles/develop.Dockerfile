FROM golang:1.24.1-bullseye

WORKDIR /app
RUN apt-get update  \
 && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    ffmpeg \
    libopus-dev \
  &&  \
    apt-get clean \
  && \
    rm -rf /var/lib/apt/lists/*
RUN go install github.com/cosmtrek/air@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . ./

ENTRYPOINT [ "make", "air" ]

