---
version: v1.1.0
---

# Using Helmsman from a docker image on Windows 10

If you have Windows 10 with Docker installed, you might be able to run Helmsman in a linux container on Windows.

1. Switch to the Linux containers from the docker tray icon.
2. Configure your local kubectl on Windows to connect to your cluster.
3. Configure your desired state file to use the kubeContext only. i.e. no cluster connection settings.
2. Run the following command:

```
docker run --rm -it -v <your kubectl config location>:/root/.kube -v <your dsf.toml directory>:/tmp  praqma/helmsman:v1.0.2 helmsman -f dsf.toml --debug --apply
```

