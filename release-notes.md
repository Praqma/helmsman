# v1.1.0

This release introduces some new features and comes with several enhancements in logging and validation.

- Introducing `priority` key for apps in the desired state to define the priority (order) of processing apps (useful for dependencies).  
- Introducing `wait` key for apps to block helmsman execution until a release operation is complete.
- Intorducing the `--ns-override` flag for overriding namespaces defined in the desired state (useful for deploying from git branches to namespaces).
- Support initializing helm with a k8s service account.
- Introducing the `--verbose` flag for more detailed logs.
- Cleaning up any downloaded certs/keys after execution.
- Improved logging with full helm error messages.
- Improved validations for desired states.
- Bumping Helm and Kubectl versions in docker images.
- Providing multiple docker image tags for different helm versions.
- Fixing not waiting for helm Tiller to be ready.
- Fixing a bug in helm release search.