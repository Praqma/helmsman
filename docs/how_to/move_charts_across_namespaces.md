---
version: v0.1.3
---

# move charts across namespaces

If you have a workflow for testing a release first in the `staging` namespace then move to the `production` namespace, Helmsman can help you.

```
...

[namespaces]
staging = "staging" 
production = "default"
myOtherNamespace = "namespaceX"

[apps]

    [apps.jenkins]
    name = "jenkins" 
    description = "jenkins"
    namespace = "staging" # this is where it is deployed
    enabled = true 
    chart = "stable/jenkins" 
    version = "0.9.1" 
    valuesFile = "" 
    purge = false 
    test = true  

...
``` 

Then if you change the namespace key for jenkins:

```
...

[namespaces]
staging = "staging" 
production = "default"
myOtherNamespace = "namespaceX"

[apps]

    [apps.jenkins]
    name = "jenkins" 
    description = "jenkins"
    namespace = "production" # we want to move it to production
    enabled = true 
    chart = "stable/jenkins" 
    version = "0.9.1" 
    valuesFile = "" 
    purge = false 
    test = true  

...
``` 

Helmsman will delete the jenkins release from the `staging` namespace and install it in the `production` namespace (default in the above setup).