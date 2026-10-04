FROM node:24-alpine AS ui-builder

WORKDIR /app

COPY chud-ui/package*.json ./
RUN npm ci

COPY chud-ui/ .

RUN npm run build

FROM golang:1.27.1-alpine AS server-builder

ARG VERSION
ARG GIT_COMMIT
ARG CI_BUILD_NUMBER
ARG IMAGE_TAG

WORKDIR /app

COPY chud-server/src/go.mod chud-server/src/go.sum* ./

RUN go mod download

COPY chud-server/src/ ./src

COPY --from=ui-builder /app/dist ./src/static

WORKDIR /app/src

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w \
    -X github.com/sebnow/chud/platform/meta.CIBuildNumber=$CI_BUILD_NUMBER \
    -X github.com/sebnow/chud/platform/meta.GitCommit=$GIT_COMMIT \
    -X github.com/sebnow/chud/platform/meta.ImageTag=$IMAGE_TAG \
    -X github.com/sebnow/chud/platform/meta.Version=$VERSION" \
    -o main .

RUN mkdir -p /app/tmp

FROM scratch

COPY --from=server-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=server-builder --chown=65532:65532 /app/src/main /app/main
COPY --from=server-builder --chown=65532:65532 /app/tmp /tmp

WORKDIR /app
USER 65532:65532

EXPOSE 2137

CMD ["./main"]
