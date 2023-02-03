FROM golang:1.19-stretch AS metrics_builder

ENV CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build .

FROM alpine:3.11.0

COPY --from=metrics_builder /app/main ./

RUN addgroup -g 1000 -S appgroup \
  && adduser -u 1000 -S appuser -G appgroup

RUN chown -R appuser:appgroup /app

USER 1000

EXPOSE 8080

ENTRYPOINT [ "/app/main" ]

