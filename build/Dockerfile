ARG GO_VERSION=1.19

#######################
## Go dependencies
## Speed up this step by mounting your local go mod pkg directory
#######################
FROM golang:${GO_VERSION}-alpine as artifact
RUN apk --no-cache add build-base git make gcc musl-dev linux-headers
WORKDIR /go/src/github.com/dennigogo/zitadel
ADD . .
RUN make build

FROM alpine:3 as final
RUN apk --no-cache add bash
WORKDIR /app
COPY --from=artifact /etc/passwd /etc/passwd
COPY --from=artifact /etc/ssl/certs /etc/ssl/certs
COPY --from=artifact /go/src/github.com/dennigogo/zitadel/zitadel .
COPY --from=artifact /go/src/github.com/dennigogo/zitadel/console .
RUN chmod +x /app/zitadel
ENV PATH="/app":$PATH
HEALTHCHECK NONE
ENTRYPOINT ["zitadel"]