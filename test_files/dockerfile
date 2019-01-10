#  This is a docker image for the helmsman test container
#  It can be pulled from praqma/helmsman-test

ARG KUBE_VERSION
ARG HELM_VERSION

FROM golang:1.10-alpine3.7 

ENV KUBE_VERSION ${KUBE_VERSION:-v1.11.3}
ENV HELM_VERSION ${HELM_VERSION:-v2.11.0}

RUN apk --no-cache update \
    && apk add --update --no-cache ca-certificates git \
    && apk add --update -t deps curl tar gzip make bash \
    && rm -rf /var/cache/apk/* \
    && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
    && chmod +x /usr/local/bin/kubectl \
    && curl -L http://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz | tar zxv -C /tmp \
    && mv /tmp/linux-amd64/helm /usr/local/bin/helm \
    && rm -rf /tmp/linux-amd64 \
    && chmod +x /usr/local/bin/helm

WORKDIR src/helmsman

RUN mkdir -p ~/.helm/plugins \
    && helm plugin install https://github.com/hypnoglow/helm-s3.git \
    && helm plugin install https://github.com/nouney/helm-gcs \
    && helm plugin install https://github.com/databus23/helm-diff \
    && helm plugin install https://github.com/futuresimple/helm-secrets \
    && rm -r /tmp/helm-diff /tmp/helm-diff.tgz

RUN go get github.com/goreleaser/goreleaser && \
    go get github.com/golang/dep/cmd/dep
