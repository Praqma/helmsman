---
version: v1.5.0
---

# install releases

You can run helmsman with the [example.toml](https://github.com/Praqma/helmsman/blob/master/example.toml) or [example.yaml](https://github.com/Praqma/helmsman/blob/master/example.yaml) file.

```

$ helmsman --apply -f example.toml
2017/11/19 18:17:57 Parsed [[ example.toml ]] successfully and found [ 2 ] apps.
2017/11/19 18:17:59 WARN: I could not create namespace [staging ]. It already exists. I am skipping this.
2017/11/19 18:17:59 WARN: I could not create namespace [default ]. It already exists. I am skipping this.
2017/11/19 18:18:02 INFO: Executing the following plan ...
---------------
Ok, I have generated a plan for you at: 2017-11-19 18:17:59.347859706 +0100 CET m=+2.255430021
DECISION: release [ jenkins ] is not present in the current k8s context. Will install it in namespace [[ staging ]]
DECISION: release [ artifactory ] is not present in the current k8s context. Will install it in namespace [[ staging ]]
2017/11/19 18:18:02 INFO: attempting: --   installing release [ jenkins ] in namespace [[ staging ]]
2017/11/19 18:18:05 INFO: attempting: --   installing release [ artifactory ] in namespace [[ staging ]]

```

```
$ helm list --namespace staging
NAME       	REVISION	UPDATED                 	STATUS  	CHART            	NAMESPACE
artifactory	1       	Sun Nov 19 18:18:06 2017	DEPLOYED	artifactory-6.2.0	staging
jenkins    	1       	Sun Nov 19 18:18:03 2017	DEPLOYED	jenkins-0.9.1    	staging
```

# delete releases

You can then change your desire, for example to disable the Jenkins release that was created above by setting `enabled = false` :

Then run Helmsman again and it will detect that you want to delete Jenkins:

> Note: As of v1.4.0-rc, deleting the jenkins app entry in the desired state file WILL result in deleting the jenkins release. To prevent this, use the `--keep-untracked-releases` flag with your Helmsman command.

```
$ helmsman --apply -f example.toml
2017/11/19 18:28:27 Parsed [[ example.toml ]] successfully and found [ 2 ] apps.
2017/11/19 18:28:29 WARN: I could not create namespace [staging ]. It already exists. I am skipping this.
2017/11/19 18:28:29 WARN: I could not create namespace [default ]. It already exists. I am skipping this.
2017/11/19 18:29:01 INFO: Executing the following plan ...
---------------
Ok, I have generated a plan for you at: 2017-11-19 18:28:29.437061909 +0100 CET m=+1.987623555
DECISION: release [ jenkins ] is desired to be deleted . Planning this for you!
DECISION: release [ artifactory ] is desired to be upgraded. Planning this for you!
2017/11/19 18:29:01 INFO: attempting: --   deleting release [ jenkins ]
2017/11/19 18:29:11 INFO: attempting: --   upgrading release [ artifactory ]
```

```
$ helm list --namespace staging
NAME       	REVISION	UPDATED                 	STATUS  	CHART            	NAMESPACE
artifactory	2       	Sun Nov 19 18:29:11 2017	DEPLOYED	artifactory-6.2.0	staging
```

If you would like the release to be deleted along with its history, you can use the `purge` flag in your desired state file as follows:

> NOTE: purge deleting a release means you can't roll it back.

```toml
...
[apps]

    [apps.jenkins]
    name = "jenkins"
    description = "jenkins"
    namespace = "staging"
    enabled = false # this tells helmsman to delete it
    chart = "stable/jenkins"
    version = "0.9.1"
    valuesFile = ""
    purge = true # this means purge delete this release whenever it is required to be deleted
    test = false

...
```

```yaml
...
apps:
  jenkins:
    name: "jenkins"
    description: "jenkins"
    namespace: "staging"
    enabled: false # this tells helmsman to delete it
    chart: "stable/jenkins"
    version: "0.9.1"
    valuesFile: ""
    purge: true # this means purge delete this release whenever it is required to be deleted
    test: false

...
```

# rollback releases

> Rollbacks in helm versions 2.8.2 and higher may not work due to a [bug](https://github.com/helm/helm/issues/3722).
Similarly, if you change `enabled` back to `true`, it will figure out that you would like to roll it back.

```
$ helmsman --apply -f example.toml
2017/11/19 18:30:41 Parsed [[ example.toml ]] successfully and found [ 2 ] apps.
2017/11/19 18:30:42 WARN: I could not create namespace [staging ]. It already exists. I am skipping this.
2017/11/19 18:30:43 WARN: I could not create namespace [default ]. It already exists. I am skipping this.
2017/11/19 18:30:49 INFO: Executing the following plan ...
---------------
Ok, I have generated a plan for you at: 2017-11-19 18:30:43.108693039 +0100 CET m=+1.978435517
DECISION: release [ jenkins ] is currently deleted and is desired to be rolledback to namespace [[ staging ]] . No problem!
DECISION: release [ artifactory ] is desired to be upgraded. Planning this for you!
2017/11/19 18:30:49 INFO: attempting: --   rolling back release [ jenkins ]
2017/11/19 18:30:50 INFO: attempting: --   upgrading release [ artifactory ]
```

# upgrade releases

Every time you run Helmsman, (unless the release is [protected or deployed in a protected namespace](protect_namespaces_and_releases.md)) it will upgrade existing deployed releases to the version you specified in the desired state file. It also applies the `values.yaml` file you specify with each install/upgrade. This means that when you don't change anything for a specific release, Helmsman would upgrade with the `values.yaml` file you provide (just in case it is a new file or you changed something there.)

If you change the chart, the existing release will be deleted and a new one with the same name will be created using the new chart.


