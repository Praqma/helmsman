# What is Helmsman?

Helmsman is a tool which adds another layer of abstraction on top of Helm (the Kubernetes package manager). It allows you to automate the deployment/management of your Helm charts (k8s packaged applications).

# How does it work?

Helmsman uses a simple declarative TOML file to allow you to describe a desired state for your applications as in the example below:

```
[settings]
kubeContext = "minikube"

# metadata -- add as many key/value pairs as you want
[metadata]
org = "orgX"
maintainer = "k8s-admin"

# define your environments and thier k8s namespaces
# syntax: environment_name = "k8s_namespace"
[namespaces]
staging = "staging" 
production = "default"


# define any private/public helm charts repos you would like to get charts from
# syntax: repo_name = "repo_url"
[helmRepos]
stable = "https://kubernetes-charts.storage.googleapis.com"
incubator = "http://storage.googleapis.com/kubernetes-charts-incubator"

# define the desired state of your applications helm charts
# each contains the following:

[apps]

    [apps.jenkins]
    name = "jenkins" # should be unique across all apps
    description = "first Jira deployment"
    env = "staging" # maps to the namespace as defined in environmetns above
    enabled = true # change to false if you want to delete this app release [empty = flase]
    chart = "stable/jenkins" # changing the chart name means delete and recreate this chart
    version = "0.9.0"
    valuesFile = "values.yaml" # from this TOML file
    purge = false # will only be considered when there is a delete operation

``` 

From the above file, Helmsman sees what you desire, validate that your desire makes sense (e.g. that the chart you desire is available inthe repos you defined) compare it with the current state of Helm and figure out what to do to make your desire true.

``` 
$ helmsman -f example.toml -apply
2017/11/02 20:32:00 INFO: executing command: helm 
2017/11/02 20:32:00 Parsed [[ example.toml ]] successfully and found [1] apps
2017/11/02 20:32:00 INFO: executing command: kubectl config use-context minikube
2017/11/02 20:32:00 INFO: executing command: helm repo add stable https://kubernetes-charts.storage.googleapis.com
2017/11/02 20:32:01 INFO: executing command: helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
2017/11/02 20:32:01 INFO: executing command: helm search stable/jenkins --version 0.9.0
2017/11/02 20:32:01 INFO: executing command: kubectl create namespace default
2017/11/02 20:32:01 WARN: I could not create namespace [default ]. It already exists. I am skipping this.
2017/11/02 20:32:01 INFO: executing command: kubectl create namespace staging
2017/11/02 20:32:01 WARN: I could not create namespace [staging ]. It already exists. I am skipping this.
2017/11/02 20:32:01 INFO: executing command: helm list --deployed -q --namespace staging
2017/11/02 20:32:02 INFO: executing command: helm list --deleted -q
2017/11/02 20:32:03 INFO: executing command: helm list --all -q
2017/11/02 20:32:03 DECISION: release [ jenkins ] is not present in the current k8s context. Will install it in namespace [[ staging ]]
2017/11/02 20:32:03 INFO: Executing the following plan ... 
Printing the current plan which was generated at: 2017-11-02 20:32:01.983165117 +0100 CET m=+1.840723078 
DECISION: release [ jenkins ] is not present in the current k8s context. Will install it in namespace [[ staging ]]
2017/11/02 20:32:03 INFO: attempting: --   installing release [ jenkins ] in namespace [[ staging ]]
2017/11/02 20:32:03 INFO: executing command: helm install stable/jenkins -n jenkins --namespace staging -f values.yaml

``` 

```
helm list
NAME        	REVISION	UPDATED                 	STATUS  	CHART        	NAMESPACE
jenkins     	1       	Thu Nov  2 20:32:05 2017	DEPLOYED	jenkins-0.9.0	staging   
```

You can then change your desire, for example to remove the Jenkins release we created above:

```
...
[apps.jenkins]
    name = "jenkins" # should be unique across all apps
    description = "first Jira deployment"
    env = "staging" # maps to the namespace as defined in environmetns above
    enabled = false # change to false if you want to delete this app release [empty = flase]
    chart = "stable/jenkins" # changing the chart name means delete and recreate this chart
    version = "0.9.0"
    valuesFile = "values.yaml" # from this TOML file
    purge = false # will only be considered when there is a delete operation
...

```




```
helmsman -f example.toml -apply
2017/11/02 20:42:35 INFO: executing command: helm 
2017/11/02 20:42:36 Parsed [[ example.toml ]] successfully and found [1] apps
2017/11/02 20:42:36 INFO: executing command: kubectl config use-context minikube
2017/11/02 20:42:36 INFO: executing command: helm repo add stable https://kubernetes-charts.storage.googleapis.com
2017/11/02 20:42:37 INFO: executing command: helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
2017/11/02 20:42:37 INFO: executing command: helm search stable/jenkins --version 0.9.0
2017/11/02 20:42:37 INFO: executing command: kubectl create namespace staging
2017/11/02 20:42:37 WARN: I could not create namespace [staging ]. It already exists. I am skipping this.
2017/11/02 20:42:37 INFO: executing command: kubectl create namespace default
2017/11/02 20:42:37 WARN: I could not create namespace [default ]. It already exists. I am skipping this.
2017/11/02 20:42:37 INFO: executing command: helm list --deployed -q --namespace staging
2017/11/02 20:42:38 DECISION: release [ jenkins ] is desired to be deleted . Planing this for you!
2017/11/02 20:42:38 INFO: Executing the following plan ... 
Printing the current plan which was generated at: 2017-11-02 20:42:37.851080895 +0100 CET m=+1.857071023 
DECISION: release [ jenkins ] is desired to be deleted . Planing this for you!
2017/11/02 20:42:38 INFO: attempting: --   deleting release [ jenkins ]
2017/11/02 20:42:38 INFO: executing command: helm delete  jenkins
```

Now, we can change the `enabled` flag for Jenkins again to ture and Helmsman will understand that we want to rollback our deleted Jenkins.

```
helmsman -f example.toml -apply
2017/11/02 20:46:36 INFO: executing command: helm 
2017/11/02 20:46:36 Parsed [[ example.toml ]] successfully and found [1] apps
2017/11/02 20:46:36 INFO: executing command: kubectl config use-context minikube
2017/11/02 20:46:36 INFO: executing command: helm repo add stable https://kubernetes-charts.storage.googleapis.com
2017/11/02 20:46:38 INFO: executing command: helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
2017/11/02 20:46:38 INFO: executing command: helm repo add novelda s3://novelda-helm-repo/charts
2017/11/02 20:46:39 INFO: executing command: helm search stable/jenkins --version 0.9.0
2017/11/02 20:46:40 INFO: executing command: kubectl create namespace staging
2017/11/02 20:46:41 WARN: I could not create namespace [staging ]. It already exists. I am skipping this.
2017/11/02 20:46:41 INFO: executing command: kubectl create namespace default
2017/11/02 20:46:41 WARN: I could not create namespace [default ]. It already exists. I am skipping this.
2017/11/02 20:46:41 INFO: executing command: helm list --deployed -q --namespace staging
2017/11/02 20:46:43 INFO: executing command: helm list --deleted -q
2017/11/02 20:46:47 INFO: executing command: helm status jenkins
2017/11/02 20:46:51 DECISION: release [ jenkins ] is currently deleted and is desired to be rolledback to namespace [[ staging ]] . No problem!
2017/11/02 20:46:51 INFO: Executing the following plan ... 
Printing the current plan which was generated at: 2017-11-02 20:46:41.793215409 +0100 CET m=+5.703781381 
DECISION: release [ jenkins ] is currently deleted and is desired to be rolledback to namespace [[ staging ]] . No problem!
2017/11/02 20:46:51 INFO: attempting: --   rolling back release [ jenkins ]
2017/11/02 20:46:51 INFO: executing command: helm rollback jenkins
```