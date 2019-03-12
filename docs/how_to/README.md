---
version: v1.8.0
---

# How To Guides

This page contains a list of guides on how to use Helmsman.

- Connecting to Kubernetes clusters
    - [Using an existing kube context](settings/existing_kube_context.md)
    - [Using the current kube context](settings/current_kube_context.md)
    - [Connecting with certificates](settings/creating_kube_context_with_certs.md)
    - [Connecting with bearer token](settings/creating_kube_context_with_token.md)
- Defining Namespaces    
    - [Create namespaces](namespaces/create.md)
    - [Label namespaces](namespaces/labels_and_annotations.md)
    - [Set resource limits for namespaces](namespaces/limits.md)
    - [Protecting namespaces](namespaces/protection.md)
- Deploying Helm Tiller
    - [Using existing Tillers](tiller/existing.md) 
    - [Deploy shared Tiller in kube-system](tiller/shared.md)
    - [Prevent Deploying Tiller in kube-system](tiller/prevent_tiller_in_kube_system.md)
    - [Deploy Multiple Tillers with custom setup for each](tiller/multitenancy.md)
    - [Deploy apps with specific Tillers](deploy_apps_with_specific_tiller.md)
- Defining Helm repositories
    - [Using default helm repos](helm_repos/default.md) 
    - [Using private repos in Google GCS](helm_repos/gcs.md)
    - [Using private repos in AWS S3](helm_repos/s3.md)    
    - [Using private repos with basic auth](helm_repos/basic_auth.md)
    - [Using pre-configured repos](helm_repos/pre_configured.md) 
    - [Using local charts](helm_repos/local.md)
- Manipulating Apps
    - [Basic operations](apps/basic.md)
    - [Passing secrets from env vars](apps/secrets.md)
    - [Use multiple values files for apps](apps/multiple_values_files.md)
    - [Protect releases (apps)](apps/protection.md)
    - [Moving releases (apps) across namespaces](apps/moving_across_namespaces.md)
    - [Override defined namespaces](apps/override_namespaces.md)
    - [Run helm tests for deployed releases (apps)](apps/helm_tests.md)  
    - [Define the order of apps operations](apps/order.md)  
    - [Delete all releases (apps)](apps/destroy.md)
- Running Helmsman in different environments
    - [Running Helmsman in CI](deployment/ci.md)
    - [Running Helmsman inside your k8s cluster](inside_k8s.md)
- Misc
    - [Authenticating to cloud storage providers](auth_to_storage_providers.md)
    - [Send slack notifications from Helmsman](send_slack_notifications_from_helmsman.md)
    - [Merge multiple desired state files](merge_desired_state_files.md)
    - [Multitenant clusters guide](multitenant_clusters_guide.md)
    - [Helmsman on Windows 10](helmsman_on_windows10.md)
