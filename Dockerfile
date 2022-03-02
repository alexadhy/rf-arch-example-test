
FROM golang:1.17-alpine as builder

ARG build_opts

COPY . /go/src/app
WORKDIR /go/src/app

RUN apk add --no-cache git openssh && \
    chmod 600 /go/src/app/repo-key && \
    echo "IdentityFile /go/src/app/repo-key" >> /etc/ssh/ssh_config && \
    echo -e "StrictHostKeyChecking no" >> /etc/ssh/ssh_config && \
    git config --global uri."ssh://git@github.com/".insteadOf "https://github.com/" && \
    export GOPRIVATE=github.com/alexadhy/* && \
    go mod tidy && \
    go build "${build_opts}" -o /app /go/src/app

RUN ls /

FROM alpine:latest

WORKDIR /

COPY --from=builder /app /app
RUN apk --no-cache add dumb-init

CMD /app
