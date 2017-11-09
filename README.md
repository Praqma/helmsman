# What is Helmsman?

Helmsman is a Helm Charts as Code tool which adds another layer of abstraction on top of [Helm](https://helm.sh) (the [Kubernetes](https://kubernetes.io/) package manager). It allows you to automate the deployment/management of your Helm charts (k8s packaged applications).

# Why Helmsman?

Helmsman was created to ease continous deployment of Helm charts. When you want to configure a continous deployment pipeline to manage multiple charts deployed on your k8s cluster(s), a CI script will quickly become complex and difficult to maintain. That's where Helmsman comes to rescue. Read more about [how Helmsman can save you time and effort in the docs](docs/why_helmsman.md).  


# How does it work?

Helmsman uses a simple declarative [TOML](https://github.com/toml-lang/toml) file to allow you to describe a desired state for your k8s applications as in the [example file](example.toml).

The desired state file follows the [desired state specification](docs/desired_state_specification.md).

Helmsman sees what you desire, validates that your desire makes sense (e.g. that the charts you desire are available in the repos you defined), compares it with the current state of Helm and figures out what to do to make your desire come true. Below is the result of executing the [example.toml](example.toml)

``` 
$ helmsman -f example.toml -apply
2017/11/04 17:23:34 Parsed [[ example.toml ]] successfully and found [2] apps
2017/11/04 17:23:49 WARN: I could not create namespace [staging ]. It already exists. I am skipping this.
2017/11/04 17:23:49 WARN: I could not create namespace [default ]. It already exists. I am skipping this.
---------------
Ok, I have generated a plan for you at: 2017-11-04 17:23:49.649943386 +0100 CET m=+14.976742294 
DECISION: release [ jenkins ] is currently deleted and is desired to be rolledback to namespace [[ staging ]] . No problem!
DECISION: release [ jenkins ] is required to be tested when installed/upgraded/rolledback. Got it!
DECISION: release [ vault ] is not present in the current k8s context. Will install it in namespace [[ staging ]]
DECISION: release [ vault ] is required to be tested when installed/upgraded/rolledback. Got it!
``` 

```
$ helm list
NAME        	REVISION	UPDATED                 	STATUS  	CHART        	NAMESPACE
jenkins     	1       	Thu Nov  4 17:24:05 2017	DEPLOYED	jenkins-0.9.0	staging 
vault        	1       	Thu Nov  4 17:24:55 2017	DEPLOYED	vault-0.1.0 	staging   
```

You can then change your desire, for example to disable the Jenkins release that was created above by setting `enabled = false` :

```
...
[apps.jenkins]
    name = "jenkins" # should be unique across all apps
    description = "jenkins"
    env = "staging" # maps to the namespace as defined in environmetns above
    enabled = false # change to false if you want to delete this app release [empty = flase]
    chart = "stable/jenkins" # changing the chart name means delete and recreate this chart
    version = "0.9.0"
    valuesFile = "" # leaving it empty uses the default chart values
    purge = false # will only be considered when there is a delete operation
    test = true # run the tests whenever this release is installed/upgraded/rolledback
...

```

Then run Helmsman again and it will detect that you want to delete Jenkins:

```
$ helmsman -f example.toml -apply
2017/11/04 17:25:29 Parsed [[ example.toml ]] successfully and found [2] apps
2017/11/04 17:25:44 WARN: I could not create namespace [staging ]. It already exists. I am skipping this.
2017/11/04 17:25:44 WARN: I could not create namespace [default ]. It already exists. I am skipping this.
---------------
Ok, I have generated a plan for you at: 2017-11-04 17:23:44.649947467 +0100 CET m=+14.976746752
DECISION: release [ jenkins ] is desired to be deleted and purged!. Planing this for you!
```

```
$ helm list
NAME        	REVISION	UPDATED                 	STATUS  	CHART        	NAMESPACE
vault        	1       	Thu Nov  4 17:24:55 2017	DEPLOYED	vault-0.1.0 	staging 
```

Similarly, if you change `enabled` back to `true`, it will figure out that you would like to roll it back. You can also change the chart or chart version and specify a values.yaml file to override the default chart values.

# Usage

Helmsman can be used in two ways:

1. In a continuous deployment pipeline. Helmsman can be used in a docker container run by your CI system to maintain your desired state (which you can store in a version control repository). The docker image is available on [dockerhub](https://hub.docker.com/r/praqma/helmsman/). 

```
docker run -it --rm -v /local/path/to/your/desired_state_file:/tmp praqma/helmsman  -f tmp/example.toml  
```
> The latest docker image will contain the latest build of Helmsman. 

2. As a binary application. Helmsman dependes on [Helm](https://helm.sh) and [Kubectl](https://kubernetes.io/docs/user-guide/kubectl/) being installed. See below for installation.

# Installation 

Install Helmsman for your OS from the [releases page](https://github.com/Praqma/Helmsman/releases). Available for Linux and MacOS.

# Documentaion

Documentation can be found under the [docs](/docs/) directory.

# Contributing
Contribution and feeback/feature requests are welcome. Please check the [Contribution Guide](CONTRIBUTING.md).