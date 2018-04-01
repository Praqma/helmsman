---
version: v1.1.0
---

# define namespaces

You can define namespaces to be used in your cluster. If they don't exist, Helmsman will create them for you.

```
...

[namespaces]
[namespaces.staging]
[namespaces.production]
  protected = true # default is false
...
``` 

You can then tell Helmsman to put specific releases in a specific namespace:

```
...
[apps]

    [apps.jenkins]
    name = "jenkins" 
    description = "jenkins"
    namespace = "myOtherNamespace" # this is the pointer to the namespace defined above -- i.e. it deploys to namespace 'namespaceX'
    enabled = true 
    chart = "stable/jenkins" 
    version = "0.9.1" 
    valuesFile = "" 
    purge = false 
    test = true  

...
``` 