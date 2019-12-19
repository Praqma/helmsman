ARG GO_VERSION="1.13.5"
ARG ALPINE_VERSION="3.10"
ARG GLOBAL_KUBE_VERSION="v1.14.8"
ARG GLOBAL_HELM_VERSION="v3.0.2"
ARG GLOBAL_HELM_DIFF_VERSION="v3.0.0-rc.7"


FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as builder
ARG GLOBAL_KUBE_VERSION
ARG GLOBAL_HELM_VERSION
ARG GLOBAL_HELM_DIFF_VERSION
ENV KUBE_VERSION=$GLOBAL_KUBE_VERSION
ENV HELM_VERSION=$GLOBAL_HELM_VERSION
ENV HELM_DIFF_VERSION=$GLOBAL_HELM_DIFF_VERSION
WORKDIR /go/src/github.com/Praqma/helmsman
COPY scripts/ /tmp/
RUN sh /tmp/setup.sh \
    && apk --no-cache add dep
COPY . .
RUN make test \
    && LastTag=$(git describe --abbrev=0 --tags) \
    && TAG=$LastTag-$(date +"%d%m%y") \
    && LT_SHA=$(git rev-parse ${LastTag}^{}) \
    && LC_SHA=$(git rev-parse HEAD) \
    && if [ ${LT_SHA} != ${LC_SHA} ]; then TAG=latest-$(date +"%d%m%y"); fi \
    && make build


FROM alpine:${ALPINE_VERSION} as base
ARG GLOBAL_KUBE_VERSION
ARG GLOBAL_HELM_VERSION
ARG GLOBAL_HELM_DIFF_VERSION
ENV KUBE_VERSION=$GLOBAL_KUBE_VERSION
ENV HELM_VERSION=$GLOBAL_HELM_VERSION
ENV HELM_DIFF_VERSION=$GLOBAL_HELM_DIFF_VERSION
COPY scripts/ /tmp/
RUN sh /tmp/setup.sh
COPY --from=builder /go/src/github.com/Praqma/helmsman/helmsman /bin/helmsman
