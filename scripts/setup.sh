#!/bin/bash

set -e

apk add --update --no-cache ca-certificates git openssh ruby curl tar gzip make bash gnupg
curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl
chmod +x /usr/local/bin/kubectl

curl -L https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz | tar zxv -C /tmp
mv /tmp/linux-amd64/helm /usr/local/bin/helm
rm -rf /tmp/linux-amd64
chmod +x /usr/local/bin/helm

mkdir -p ~/.helm/plugins
helm plugin install https://github.com/hypnoglow/helm-s3.git
helm plugin install https://github.com/nouney/helm-gcs
helm plugin install https://github.com/databus23/helm-diff --version ${HELM_DIFF_VERSION}
helm plugin install https://github.com/futuresimple/helm-secrets
rm -r /tmp/helm-diff /tmp/helm-diff.tgz
gem install hiera-eyaml --no-doc
