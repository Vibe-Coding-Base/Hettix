ARG GO_VERSION=1.26
ARG NODE_VERSION=22
ARG ALPINE_VERSION=3.21

FROM node:${NODE_VERSION}-alpine AS node-builder
WORKDIR /app
COPY admin/package.json admin/yarn.lock ./
RUN yarn install --frozen-lockfile
COPY admin/ .
RUN yarn run build

FROM golang:${GO_VERSION}-alpine AS go-builder
ARG HETTIX_VERSION=0.0.0
ENV CGO_ENABLED=0
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY pkg ./pkg
COPY --from=node-builder /app/dist ./pkg/adminui/admin
RUN go build -ldflags="-s -w -X main.version=${HETTIX_VERSION}" -o /hettix ./cmd/hettix

FROM alpine:${ALPINE_VERSION}
WORKDIR /app
COPY --from=go-builder /hettix .

ENTRYPOINT ["./hettix"]

EXPOSE 8080
