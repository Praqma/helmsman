---
version: v0.1.2
---

# Usage

```
docker run -v $(pwd):/tmp --rm -it \
-e KUBECTL_PASSWORD=<k8s_password> \
-e AWS_ACCESS_KEY_ID=<aws_key_id> \
-e AWS_DEFAULT_REGION=<aws_region> \
-e AWS_SECRET_ACCESS_KEY=<acess_key> \
praqma/helmsman:v0.1.2 \
helmsman -debug -apply -f <your_desired_state_file>.<toml|yaml>
```

Check the different image tags on [Dockerhub](https://hub.docker.com/r/praqma/helmsman/)