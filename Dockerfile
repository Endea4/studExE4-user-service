FROM golang:1.25-alpine AS builder
ARG GITHUB_TOKEN
ARG GOPRIVATE=github.com/Endea4/*
RUN apk add --no-cache git ca-certificates
RUN git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
ENV GOPRIVATE=${GOPRIVATE}

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /user-service ./cmd/server

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=builder /user-service /user-service
EXPOSE 9086
CMD ["/user-service"]
