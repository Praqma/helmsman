---
version: v1.0.0
---

# Best Practice

When using Helmsman, we recommend the following best practices:

- Add useful metadata in your desired state files (DSFs) so that others (who have access to them) can make understandable what your DSF is for. We recommend the following metadata: organization, maintainer (name and email), and description/purpose.

- Use environment variables to pass K8S connection secrets (password, certificates paths on the local system or AWS/GCS bucket urls and the API URI). This keeps all sensitive information out of your version controlled source code.

- Define certain namespaces (e.g, production) as protected namespaces (supported in v1.0.0+) and deploy your production-ready releases there.

- If you use multiple desired state files (DSFs) with the same cluster, make sure your namespace protection definitions are identical across all DSFs.

- Don't maintain the same release in multiple DSFs.

- While the decision on how many DSFs to use and what each can contain is up to you and depends on your case, we recommend coming up with your own rules for how to split them.

