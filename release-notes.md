# v0.1.3 

- Bug fixes.
- Support for passing user input values to charts from environment variables. Such values either override values from values.yaml or add to them. This is particularily useful for passing secrets from environment variables to helm charts.
- Tests no longer run on upgrade and rollback operations. This is because some charts does not use unique generated names for the tests and Helm does not manage the tests. In some cases, tests would fail due to having k8s resources with the same names existing from a previous run. 
- Support for use of certificates and keys for cluster connection from the local file system. YOU SHOULD NEVER commit your certificates to git! This update is useful while testing on a local machine.
- The key 'env' for release namespaces in the desired state file is now changed to the more sensible 'namespace'. 