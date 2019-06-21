#  This is a docker image for helmsman


FROM golang:1.11.11-alpine3.10 as builder

WORKDIR /go/src/

RUN apk --no-cache add make git
RUN git clone https://github.com/Praqma/helmsman.git

#  build a statically linked binary so that it works on stripped linux images such as alpine/busybox.
RUN cd helmsman \
    && LastTag=$(git describe --abbrev=0 --tags) \
    && TAG=$LastTag-$(date +"%d%m%y") \
    && LT_SHA=$(git rev-parse ${LastTag}^{}) \
    && LC_SHA=$(git rev-parse HEAD) \
    && if [ ${LT_SHA} != ${LC_SHA} ]; then TAG=latest-$(date +"%d%m%y"); fi \
    && make dependencies \
    && CGO_ENABLED=0 GOOS=linux go install -a -ldflags '-X main.version='$TAG' -extldflags "-static"' .


# The image to keep
FROM alpine:3.10

ARG KUBE_VERSION
ARG HELM_VERSION

ENV KUBE_VERSION ${KUBE_VERSION:-v1.11.3}
ENV HELM_VERSION ${HELM_VERSION:-v2.11.0}
ENV HELM_DIFF_VERSION ${HELM_DIFF_VERSION:-v2.11.0+5}

RUN apk --no-cache update \
    && apk add --update --no-cache ca-certificates git openssh \
    && apk add --update -t deps curl tar gzip make bash \
    && rm -rf /var/cache/apk/* \
    && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
    && chmod +x /usr/local/bin/kubectl \
    && curl -L http://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz | tar zxv -C /tmp \
    && mv /tmp/linux-amd64/helm /usr/local/bin/helm \
    && rm -rf /tmp/linux-amd64 \
    && chmod +x /usr/local/bin/helm

COPY --from=builder /go/bin/helmsman   /bin/helmsman

RUN mkdir -p ~/.helm/plugins \
    && helm plugin install https://github.com/hypnoglow/helm-s3.git \
    && helm plugin install https://github.com/nouney/helm-gcs \
    && helm plugin install https://github.com/databus23/helm-diff --version ${HELM_DIFF_VERSION} \
    && helm plugin install https://github.com/futuresimple/helm-secrets \
    && helm plugin install https://github.com/rimusz/helm-tiller \
    && rm -r /tmp/helm-diff /tmp/helm-diff.tgz

WORKDIR /tmp
# ENTRYPOINT ["/bin/helmsman"]

