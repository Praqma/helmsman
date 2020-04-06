---
version: v3.0.0
---

# CMD reference

This lists available CMD options in Helmsman:

> you can find the CMD options for the version you are using by typing: `helmsman -h` or `helmsman --help`

  `--apply`
        apply the plan directly.

  `--context-override string`
        override releases context defined in release state with this one.

  `--debug`
        show the debug execution logs and actual helm/kubectl commands. This can log secrets and should only be used for debugging purposes.

  `--verbose`
        show verbose execution logs.

  `--destroy`
        delete all deployed releases.

  `--diff-context num`
        number of lines of context to show around changes in helm diff output.

  `-p`
        max number of concurrent helm releases to run

  `--dry-run`
        apply the dry-run (do not update) option for helm commands.

  `-e value`
        file(s) to load environment variables from (default .env), may be supplied more than once.

  `-f value`
        desired state file name(s), may be supplied more than once to merge state files.

  `--force-upgrades`
        use --force when upgrading helm releases. May cause resources to be recreated.

  `--keep-untracked-releases`
        keep releases that are managed by Helmsman from the used DSFs in the command, and are no longer tracked in your desired state.

  `--kubeconfig`
        path to the kubeconfig file to use for CLI requests.

  `--migrate-context`
        Updates the context name for all apps defined in the DSF and applies Helmsman labels. Using this flag is required if you want to change context name after it has been set.

  `--no-banner`
        don't show the banner.

  `--no-color`
        don't use colors.

  `--no-env-subst`
        turn off environment substitution globally.

  `--subst-env-values`
        turn on environment substitution in values files.

  `--no-fancy`
        don't display the banner and don't use colors.

  `--no-ns`
        don't create namespaces.

  `--no-ssm-subst`
        turn off SSM parameter substitution globally.

  `--subst-ssm-values`
        turn on SSM parameter substitution in values files.

  `--ns-override string`
        override defined namespaces with this one.

  `--show-diff`
        show helm diff results. Can expose sensitive information.

  `--skip-validation`
        skip desired state validation.

  `--target`
        limit execution to specific app.

  `--group`
        limit execution to specific group of apps.

  `--update-deps`
        run 'helm dep up' for local chart

  `--v`    show the version.
