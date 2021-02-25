# Semantic Versioning Action

This action calculates the next version relying on semantic versioning.

## Github Environment Variables

- `GITHUB_SHA`
- `GITHUB_REPOSITORY`

## Branch Names

There are the prefixes we assume when bump is set to `auto`:

- `bugfix/*`, `bugfixes/*`, `hotfix\*` or `hotfixes\*` - `patch`
- `feature/*` or `features/*` - `minor`
- `major/*` - `major`

## Strategies

Let's assume the default values for the scenarios below:

- bump = "auto"
- prefix = "v"
- prerelease_id = "pre"

### Scenarios

- Not a valid source branch prefix and `auto` bump - Increments pre-release version.

    ```text
        v0.1.0 becomes v0.1.0-pre.1
        v1.5.3-pre.2 becomes v1.5.3-pre.3
    ```

- Source branch is prefixed with `bugfix/` and dest branch is `develop` and `auto` bump - Increments patch version.

    ```text
    v0.1.0 becomes v0.1.1-pre.1
    v1.5.3-pre.2 becomes v1.5.4-pre.1
    ```

- Source branch is prefixed with `hotfix/` and dest branch is `master` and `auto` bump - Increments patch version.

    ```text
    v0.1.0 becomes v0.1.1
    v1.5.3-pre.2 becomes v1.5.4
    ```

- Source branch is prefixed with `feature/` and dest branch is `develop` and `auto` bump - Increments minor version.

    ```text
    v0.1.0 becomes v0.2.0-pre.1
    v1.5.3-pre.2 becomes v1.6.0-pre.1
    ```

- Source branch is prefixed with `major/` and dest branch is `develop` and `auto` bump - Increments major version.

    ```text
    v0.1.0 becomes v1.0.0-pre.1
    v1.5.3-pre.2 becomes v2.0.0-pre.1
    ```

- Source branch is `develop` and dest branch is `master` and `auto` bump - Takes the closest tag and finalize it.

    ```text
    v0.1.0 stays v0.1.0
    v1.5.3-pre.2 becomes v1.5.3
    ```

## Inputs

### bump

**Optional** Bump strategy for semantic versioning. Can be `auto`, `major`, `minor`, `patch`. Defaults to `auto`.

### base_version

**Optional** Version to use as base for the generation, skips version bumps.

### prefix

**Optional** Prefix used to prepend the final version. Defaults to `v`.

### prerelease_id

**Optional** Text representing the pre-release identifier. Defaults to `pre`.

### main_branch_name

**Optional** The main branch name. Defaults to `master`.

### develop_branch_name

**Optional** The development branch name. Defaults to `develop`.

### dry_run

**Optional** When true only prints the calculated semantic version. Defaults to `false`.

### tag_message

**Optional** Tag message. Defaults to `auto tag`.

### auth_token

Authorization token to authenticate through the Api. Only required when dry_run is set to false.

### debug

**Optional** Enables debug mode. Defaults to false.

## Outpus

### semver_tag

The calculdated semantic version.

### previous_tag

The tag used to calculate next semantic version.

## Example usage

### Basic

Uses `auto` strategy to calculate the next semantic version.

```yaml
uses: wakatime/semver-action@v0.1.0
  with:
    auth_token: "token"
```

### Dry Run

Do not push the created tag and write to output as `semver_tag`.

```yaml
id: semver-tag
uses: wakatime/semver-action@v0.1.0
  with:
    dry_run: "true"
- name: "Created tag"
  run: echo "tag ${{ steps.semver-tag.outputs.semver_tag }}"
```
