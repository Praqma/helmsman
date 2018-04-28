# v1.2.0-rc

This release focuses on improving Helmsman latency and supporting multi-tenant clusters.

- Up to 7x faster than previous version.
- Introducing the `--skip-validation` flag which skips validating the desired state.
- Support for multi-tenant k8s clusters through:
    - Supporting deployment of Tiller in several namespaces with different service accounts
    - Supporting securing Tillers with TLS.
    - Supporting using `Secrets` as a storage background instead of configMaps.
 - Upgrading rolledback releases automatically after rollback to avoid missing changed values.
 - More concise logs.
 - Several minor enhancements.  