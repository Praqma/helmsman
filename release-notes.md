# v1.7.3-rc

> If you are already using an older version of Helmsman than v1.4.0-rc, please read the changes below carefully and follow the upgrade guide [here](docs/migrating_to_v1.4.0-rc.md)

# Fixes:

- fixing docker images helm verions with and updating dependencies. Issues: #157 #156. PR: #158 #165
- adding `batch` to the RBAC API groups. Issue: #160. PR: #162

#New features: 

- allow `json` files to be used as values files. PR #164
- adding `LimitRange` to the namespaces definitions. PR #163
- allow using current kubecontext without specifying it and adding `--kubeconfig` flag to pass a kube config file. PR #159
- allow users to pass additional helm flags when defining apps in the desired state files. PR #161
- improved visibility of decisions output.PR #146