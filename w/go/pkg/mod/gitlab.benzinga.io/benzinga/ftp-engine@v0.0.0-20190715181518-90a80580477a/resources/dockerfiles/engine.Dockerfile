# builder
FROM golang:alpine as builder

RUN apk add --no-cache git bash sed build-base

RUN mkdir -p /go/src/gitlab.benzinga.io/benzinga/ftp-engine/

WORKDIR /go/src/gitlab.benzinga.io/benzinga/ftp-engine/

COPY . .

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

RUN export GIT_HEAD=$(git rev-parse --verify HEAD) && \
    echo "Applying build tag ${GIT_HEAD:0:8}" && \
    go build -v -a -installsuffix cgo -o app \
    -ldflags "-X main.build=${GIT_HEAD:0:8} -w -extldflags '-static'" -a -tags netgo \
    /go/src/gitlab.benzinga.io/benzinga/ftp-engine/cmd/ftp-engine-worker/main.go

# actual container
FROM gcr.io/distroless/static

WORKDIR /app

COPY --from=builder /go/src/gitlab.benzinga.io/benzinga/ftp-engine/app .

EXPOSE 9000/tcp

HEALTHCHECK --interval=10s --timeout=1s --start-period=5s \
    CMD curl -f http://localhost:9000/healthz || exit 1

CMD ["./app"]

# docker build -t registry.gitlab.benzinga.io/benzinga/ftp-engine/engine . -f resources/dockerfiles/engine.Dockerfile