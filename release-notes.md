# v0.2.0

- Support reading cluster certificates from Google Cloud Storage (GCS).
- Supporting private helm repos in GCS.
- Support using client-certificates in the cluster authentication. 
- Adding a warning about PV and PVCs possible issues when moving apps across namespaces.
- Allowing certs file paths or bucket URLs and clusterURI to be passed from environment variables. 
- Fixing a bug with undefined namespaces.
- Removing the aws cli dependency. 
- Smaller docker image.