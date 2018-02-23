---
version: v0.2.0
---

# test charts

You can specifiy that you would like a chart to be tested whenever it is installed for the first time using the `test` key as follows:

```
...
[apps]

    [apps.jenkins]
    name = "jenkins" 
    description = "jenkins"
    namespace = "staging" 
    enabled = true 
    chart = "stable/jenkins" 
    version = "0.9.1" 
    valuesFile = "" 
    purge = false 
    test = true  # setting this to true, means you want the charts tests to be run on this release when it is intalled. 

...
``` 