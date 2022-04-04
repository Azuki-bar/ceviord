FROM golang:1.18.0-buster as builder

COPY ./ /app
WORKDIR /app
RUN make build

CMD /app/ceviord
