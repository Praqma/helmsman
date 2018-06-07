---
version: v1.2.0
---

![helmsman-logo](docs/images/helmsman.png)

# What is Helmsman?

Helmsman is a Helm Charts (k8s applications) as Code tool which allows you to automate the deployment/management of your Helm charts from version controlled code.

# How does it work?

Helmsman uses a simple declarative [TOML](https://github.com/toml-lang/toml) file to allow you to describe a desired state for your k8s applications as in the [example file](https://github.com/Praqma/helmsman/blob/master/example.toml).

The desired state file (DSF) follows the [desired state specification](https://github.com/Praqma/helmsman/blob/master/docs/desired_state_specification.md).

Helmsman sees what you desire, validates that your desire makes sense (e.g. that the charts you desire are available in the repos you defined), compares it with the current state of Helm and figures out what to do to make your desire come true. 

To plan without executing:
``` $ helmsman -f example.toml ```

To plan and execute the plan:
``` $ helmsman -apply -f example.toml ```

To show debugging details:
``` $ helmsman -debug -apply -f example.toml ```

# Features

- **Built for CD**: Helmsman can be used as a docker image or a binary. 
- **Applications as code**: describe your desired applications and manage them from a single version-controlled declarative file.
- **Suitable for Multitenant Clusters**: deploy Tiller in different namespaces with service accounts and TLS.
- **Easy to use**: deep knowledge of Helm CLI and Kubectl is NOT manadatory to use Helmsman. 
- **Plan, View, apply**: you can run Helmsman to generate and view a plan with/without executing it. 
- **Portable**: Helmsman can be used to manage charts deployments on any k8s cluster.
- **Protect Namespaces/Releases**: you can define certain namespaces/releases to be protected against accidental human mistakes.
- **Define the order of managing releases**: you can define the priorities at which releases are managed by helmsman (useful for dependencies).
- **Idempotency**: As long your desired state file does not change, you can execute Helmsman several times and get the same result. 
- **Continue from failures**: In the case of partial deployment due to a specific chart deployment failure, fix your helm chart and execute Helmsman again without needing to rollback the partial successes first.

# Install

## From binary

Check the [releases page](https://github.com/Praqma/Helmsman/releases) for the different versions.
```
# on Linux
curl -L https://github.com/Praqma/helmsman/releases/download/v1.2.0/helmsman_1.2.0_linux_amd64.tar.gz | tar zx
# on MacOS
curl -L https://github.com/Praqma/helmsman/releases/download/v1.2.0/helmsman_1.2.0_darwin_amd64.tar.gz | tar zx

mv helmsman /usr/local/bin/helmsman
```

## As a docker image
Check the images on [dockerhub](https://hub.docker.com/r/praqma/helmsman/tags/)

# Documentaion

Documentation and How-Tos can be found [here](https://github.com/Praqma/helmsman/blob/master/docs/).
Helmsman lets you:

- [install/delete/upgrade/rollback your helm charts from code](https://github.com/Praqma/helmsman/blob/master/docs/how_to/manipulate_apps.md).
- [work safely in a multitenant cluster](https://github.com/Praqma/helmsman/blob/master/docs/how_to/multitenant_clusters_guide.md).
- [pass secrets/user input to helm charts from environment variables](https://github.com/Praqma/helmsman/blob/master/docs/how_to/pass_secrets_from_env_variables.md).
- [test releases when they are first installed](https://github.com/Praqma/helmsman/blob/master/docs/how_to/test_charts.md).
- [use public and private helm charts](https://github.com/Praqma/helmsman/blob/master/docs/how_to/use_private_helm_charts.md).
- [use locally developed helm charts (the tar archives)](https://github.com/Praqma/helmsman/blob/master/docs/how_to/use_local_charts.md).
- [define namespaces to be used in your cluster](https://github.com/Praqma/helmsman/blob/master/docs/how_to/define_namespaces.md).
- [move charts across namespaces](https://github.com/Praqma/helmsman/blob/master/docs/how_to/move_charts_across_namespaces.md).
- [protect namespaces/releases against accidental changes](https://github.com/Praqma/helmsman/blob/master/docs/how_to/protect_namespaces_and_releases.md)
- [Define priorities at which releases are deployed/managed](https://github.com/Praqma/helmsman/blob/master/docs/how_to/use_the_priority_key.md)
- [Override the defined namespaces to deploy all releases in a specific namespace](https://github.com/Praqma/helmsman/blob/master/docs/how_to/override_defined_namespaces.md)


## Usage 

Helmsman can be used in three different settings:

- [As a binary with Minikube](https://github.com/Praqma/helmsman/blob/master/docs/how_to/run_helmsman_with_minikube.md).
- [As a binary with a hosted cluster](https://github.com/Praqma/helmsman/blob/master/docs/how_to/run_helmsman_with_hosted_cluster.md).
- [As a docker image in a CI system or local machine](https://github.com/Praqma/helmsman/blob/master/docs/how_to/run_helmsman_in_ci.md) Always use a tagged docker image from [dockerhub](https://hub.docker.com/r/praqma/helmsman/) as the `latest` image can (at times) be unstable.


# Contributing

Pull requests, feeback/feature requests are welcome. Please check our [contribution guide](CONTRIBUTION.md).
