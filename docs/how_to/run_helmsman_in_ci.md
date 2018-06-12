---
version: v1.3.0-rc
---

# Run Helmsman in CI

You can run Helmsman as a job in your CI system using the [helmsman docker image](https://hub.docker.com/r/praqma/helmsman/). 
The following example is a `config.yml` file for CircleCI but can be replicated for other CI systems.

```
version: 2
jobs:
    
    deploy-apps:
      docker:
        - image: praqma/helmsman:v1.2.0-rc
      steps:
        - checkout
        - run:
            name: Deploy Helm Packages using helmsman
            command: helmsman -debug -apply -f helmsman-deployments.toml


workflows:
  version: 2
  build:
    jobs:
      - deploy-apps
``` 

The `helmsman-deployments.toml` is your desired state file which will version controlled in your git repo.