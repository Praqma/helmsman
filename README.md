---
version: v0.1.3
---

# What is Helmsman?

Helmsman is a Helm Charts (k8s applications) as Code tool which adds a layer of abstraction on top of [Helm](https://helm.sh) (the [Kubernetes](https://kubernetes.io/) package manager). It allows you to automate the deployment/management of your Helm charts.

# Why Helmsman?

Helmsman was created to ease continous deployment of Helm charts. When you want to configure a continous deployment pipeline to manage multiple charts deployed on your k8s cluster(s), a CI script will quickly become complex and difficult to maintain. That's where Helmsman comes to rescue. Read more about [how Helmsman can save you time and effort](https://github.com/Praqma/helmsman/blob/master/docs/why_helmsman.md).  


# Features

- **Built for CD**: Helmsman can be used as a docker image or a binary. 
- **Applications as code**: describe your desired applications and manage them from a single version-controlled declarative file.
- **Easy to use**: deep knowledge of Helm CLI and Kubectl is NOT manadatory to use Helmsman. 
- **Plan, View, apply**: you can run Helmsman to generate and view a plan with/without executing it. 
- **Portable**: Helmsman can be used to manage charts deployments on any k8s cluster.
- **Idempotency**: As long your desired state file does not change, you can execute Helmsman several times and get the same result. 
- **Continue from failures**: In the case of partial deployment due to a specific chart deployment failure, fix your helm chart and execute Helmsman again without needing to rollback the partial successes first.

# Helmsman lets you:

- [install/delete/upgrade/rollback your helm charts from code](https://github.com/Praqma/helmsman/blob/master/docs/how_to/manipulate_apps.md).
- [pass secrets/user input to helm charts from environment variables](https://github.com/Praqma/helmsman/blob/master/docs/how_to/pass_secrets_from_env_variables.md).
- [test releases when they are first installed](https://github.com/Praqma/helmsman/blob/master/docs/how_to/test_charts.md).
- [use public and private helm charts](https://github.com/Praqma/helmsman/blob/master/docs/how_to/use_private_helm_charts.md).
- [use locally developed helm charts (the tar archives)](https://github.com/Praqma/helmsman/blob/master/docs/how_to/use_local_charts.md).
- [define namespaces to be used in your cluster](https://github.com/Praqma/helmsman/blob/master/docs/how_to/define_namespaces.md).
- [move charts across namespaces](https://github.com/Praqma/helmsman/blob/master/docs/how_to/move_charts_across_namespaces.md).


# Usage 

Helmsman can be used in three different settings:

- [As a binary with Minikube](https://github.com/Praqma/helmsman/blob/master/docs/how_to/run_helmsman_with_minikube.md).
- [As a binary with a hosted cluster](https://github.com/Praqma/helmsman/blob/master/docs/how_to/run_helmsman_with_hosted_cluster.md).
- [As a docker image in a CI system or local machine](https://github.com/Praqma/helmsman/blob/master/docs/how_to/run_helmsman_in_ci.md).


# How does it work?

Helmsman uses a simple declarative [TOML](https://github.com/toml-lang/toml) file to allow you to describe a desired state for your k8s applications as in the [example file](https://github.com/Praqma/helmsman/blob/master/example.toml).

The desired state file follows the [desired state specification](https://github.com/Praqma/helmsman/blob/master/docs/desired_state_specification.md).

Helmsman sees what you desire, validates that your desire makes sense (e.g. that the charts you desire are available in the repos you defined), compares it with the current state of Helm and figures out what to do to make your desire come true. 

To plan without executing:
``` $ helmsman -f example.toml ```

To plan and execute the plan:
``` $ helmsman -apply -f example.toml ```

To show debugging details:
``` $ helmsman -debug -apply -f example.toml ```


# Installation 

Install Helmsman for your OS from the [releases page](https://github.com/Praqma/Helmsman/releases). Available for Linux and MacOS.

Also available as a [docker image](https://hub.docker.com/r/praqma/helmsman/).

# Documentaion

Documentation and How-Tos can be found [here](https://github.com/Praqma/helmsman/blob/master/docs/).

# Contributing

Contribution and feeback/feature requests are welcome. 