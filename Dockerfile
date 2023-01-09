FROM golang:1.19.4-bullseye as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
ARG VERSION=snapshot
RUN make build

FROM ubuntu:22.04 as runner
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

WORKDIR /app
COPY --from=builder /app/parameter.yaml ./
COPY --from=builder /app/ceviord ./ceviord
ENTRYPOINT [ "./ceviord" ]

