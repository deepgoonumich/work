# builder
FROM golang:alpine as builder

RUN apk add --no-cache git bash sed build-base

RUN mkdir -p /go/src/gitlab.benzinga.io/benzinga/ftp-engine/

WORKDIR /go/src/gitlab.benzinga.io/benzinga/ftp-engine/

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN export GIT_HEAD=$(git rev-parse --verify HEAD) && \
    echo "Applying build tag ${GIT_HEAD:0:8}" && \
    go build -v -a -installsuffix cgo -o app \
    -ldflags "-X main.build=${GIT_HEAD:0:8}" \
    /go/src/gitlab.benzinga.io/benzinga/ftp-engine/cmd/ftp-engine-updater/main.go

# actual container
FROM gcr.io/distroless/static

WORKDIR /app

COPY --from=builder /go/src/gitlab.benzinga.io/benzinga/ftp-engine/app .

CMD ["./app"]

# docker build -t registry.gitlab.benzinga.io/benzinga/ftp-engine/updater . -f resources/dockerfiles/updater.Dockerfile