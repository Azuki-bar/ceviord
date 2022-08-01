FROM golang:1.18.5-bullseye as builder

COPY ./ /app
WORKDIR /app
RUN make build

FROM ubuntu:jammy-20220428 as runner
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
COPY --from=builder /app/parameter.yaml /app/
COPY --from=builder /app/ceviord /app/ceviord
CMD /app/ceviord
