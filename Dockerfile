ARG GO_VERSION="1.18.5"
ARG ALPINE_VERSION="3.16"
ARG GLOBAL_KUBE_VERSION="v1.24.10"
ARG GLOBAL_HELM_VERSION="v3.11.0"
ARG GLOBAL_HELM_DIFF_VERSION="v3.6.0"
ARG GLOBAL_HELM_GCS_VERSION="0.3.21"
ARG GLOBAL_HELM_S3_VERSION="v0.10.0"
ARG GLOBAL_HELM_SECRETS_VERSION="v3.13.0"
ARG GLOBAL_SOPS_VERSION="v3.7.3"

### Helm Installer ###
FROM alpine:${ALPINE_VERSION} as helm-installer
ARG GLOBAL_KUBE_VERSION
ARG GLOBAL_HELM_VERSION
ARG GLOBAL_HELM_DIFF_VERSION
ARG GLOBAL_HELM_GCS_VERSION
ARG GLOBAL_HELM_S3_VERSION
ARG GLOBAL_HELM_SECRETS_VERSION
ARG GLOBAL_SOPS_VERSION
ENV KUBE_VERSION=$GLOBAL_KUBE_VERSION
ENV HELM_VERSION=$GLOBAL_HELM_VERSION
ENV HELM_DIFF_VERSION=$GLOBAL_HELM_DIFF_VERSION
ENV HELM_GCS_VERSION=$GLOBAL_HELM_GCS_VERSION
ENV HELM_S3_VERSION=$GLOBAL_HELM_S3_VERSION
ENV HELM_SECRETS_VERSION=$GLOBAL_HELM_SECRETS_VERSION
ENV SOPS_VERSION=$GLOBAL_SOPS_VERSION
ENV HELM_DIFF_THREE_WAY_MERGE=true

RUN apk add --update --no-cache ca-certificates git openssh-client openssl ruby curl wget tar gzip make bash

ADD https://github.com/mozilla/sops/releases/download/${SOPS_VERSION}/sops-${SOPS_VERSION}.linux /usr/local/bin/sops
RUN chmod +x /usr/local/bin/sops

RUN curl --retry 5 -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl
RUN chmod +x /usr/local/bin/kubectl

RUN curl --retry 5 -Lk https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz | tar zxv -C /tmp
RUN mv /tmp/linux-amd64/helm /usr/local/bin/helm && rm -rf /tmp/linux-amd64
RUN chmod +x /usr/local/bin/helm

RUN helm plugin install https://github.com/hypnoglow/helm-s3.git --version ${HELM_S3_VERSION}
RUN helm plugin install https://github.com/nouney/helm-gcs --version ${HELM_GCS_VERSION}
RUN helm plugin install https://github.com/databus23/helm-diff --version ${HELM_DIFF_VERSION}
RUN helm plugin install https://github.com/jkroepke/helm-secrets --version ${HELM_SECRETS_VERSION}

### Go Builder & Tester ###
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as builder

RUN apk add --update --no-cache ca-certificates git openssh-client ruby bash make curl
RUN gem install hiera-eyaml hiera-eyaml-gkms --no-doc
RUN update-ca-certificates

COPY --from=helm-installer /usr/local/bin/kubectl /usr/local/bin/kubectl
COPY --from=helm-installer /usr/local/bin/helm /usr/local/bin/helm
COPY --from=helm-installer /root/.cache/helm/plugins/ /root/.cache/helm/plugins/
COPY --from=helm-installer /root/.local/share/helm/plugins/ /root/.local/share/helm/plugins/

WORKDIR /go/src/github.com/Praqma/helmsman

COPY . .
RUN make test \
    && LastTag=$(git describe --abbrev=0 --tags) \
    && TAG=$LastTag-$(date +"%d%m%y") \
    && LT_SHA=$(git rev-parse ${LastTag}^{}) \
    && LC_SHA=$(git rev-parse HEAD) \
    && if [ ${LT_SHA} != ${LC_SHA} ]; then TAG=latest-$(date +"%d%m%y"); fi \
    && make build

### Final Image ###
FROM alpine:${ALPINE_VERSION} as base

RUN apk add --update --no-cache ca-certificates git openssh-client ruby curl bash gnupg
RUN gem install hiera-eyaml hiera-eyaml-gkms --no-doc
RUN update-ca-certificates

COPY --from=helm-installer /usr/local/bin/kubectl /usr/local/bin/kubectl
COPY --from=helm-installer /usr/local/bin/helm /usr/local/bin/helm
COPY --from=helm-installer /usr/local/bin/sops /usr/local/bin/sops
COPY --from=helm-installer /root/.cache/helm/plugins/ /root/.cache/helm/plugins/
COPY --from=helm-installer /root/.local/share/helm/plugins/ /root/.local/share/helm/plugins/

COPY --from=builder /go/src/github.com/Praqma/helmsman/helmsman /bin/helmsman
