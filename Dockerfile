# Stage 1: build
FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w \
      -X main.version=${VERSION} \
      -X main.commit=${COMMIT} \
      -X main.date=${DATE}" \
    -o /out/cultivator \
    ./cmd/cultivator

# Stage 2: runtime
FROM alpine:3.21

RUN apk add --no-cache ca-certificates git

COPY --from=builder /out/cultivator /usr/local/bin/cultivator

ENTRYPOINT ["cultivator"]
