---
version: v1.9.0
---

# CMD reference

This is the list of the available CMD options in Helmsman:

> you can find the CMD options for the version you are using by typing: `helmsman -h` or `helmsman --help`

  `--apply`
        apply the plan directly.

  `--apply-labels`
        apply Helmsman labels to Helm state for all defined apps.

  `--debug`
        show the execution logs.

  `--destroy`
        delete all deployed releases. Purge delete is used if the purge option is set to true for the releases.

  `--dry-run`
        apply the dry-run option for helm commands.

  `-e value`
        file(s) to load environment variables from (default .env), may be supplied more than once.

  `-f value`
        desired state file name(s), may be supplied more than once to merge state files.

  `--keep-untracked-releases`
        keep releases that are managed by Helmsman and are no longer tracked in your desired state.

  `--no-banner`
        don't show the banner.

  `--no-color`
        don't use colors.

  `--no-fancy`
        don't display the banner and don't use colors.

  `--no-ns`
        don't create namespaces.

  `--ns-override string`
        override defined namespaces with this one.

  `--show-diff`
        show helm diff results. Can expose sensitive information.

  `--skip-validation`
        skip desired state validation.

  `--suppress-diff-secrets`
        don't show secrets in helm diff output.

  `-v`    show the version.

  `--verbose`
        show verbose execution logs.

  `--kubeconfig`
        path to the kubeconfig file to use for CLI requests.

  `--target`
        limit execution to specific app.

  `--no-env-subst`
        turn off environment substitution globally.
