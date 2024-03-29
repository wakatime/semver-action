# Semantic Versioning Action

![Unit Tests master](https://img.shields.io/github/workflow/status/wakatime/semver-action/Unit%20Tests/master?label=%20tests) [![Coverage Status](https://coveralls.io/repos/github/wakatime/semver-action/badge.svg?branch=master)](https://coveralls.io/github/wakatime/semver-action?branch=master)

This action calculates the next version relying on semantic versioning.

## Strategies

If `auto` bump, it will try to extract the closest tag and calculate the next semantic version. If not, it will bump respecting the value passed.

### Branch Names

These are the prefixes we expect when `auto` bump:

- `^bugfix/.+` or `^hotfix/.+` - `patch`
- `^docs?/.+` - `build`
- `^feature/.+` - `minor`
- `^major/.+` - `major`
- `^misc/.+` - `build`
- `^resync/.+` - Special case needed to resync base branch into develop when hotfix gets merged into base - Mostly from `master` into `develop`.

### Scenarios

#### Auto Bump

- Not a valid source branch prefix - Increments prerelease version.

    ```text
        v0.1.0 results in v0.1.0-pre.1
        v1.5.3-pre.2 results in v1.5.3-pre.3
    ```

- Source branch is prefixed with `misc/` or `doc(s)/` and dest branch is `develop` - Increments build version.

    ```text
    v1.5.3-pre.2 results in v1.5.3-pre.3
    ```

- Source branch is prefixed with `bugfix/` and dest branch is `develop` - Increments patch version.

    ```text
    v0.1.0 results in v0.1.1-pre.1
    v1.5.3-pre.2 results in v1.5.4-pre.1
    ```

- Source branch is prefixed with `hotfix/` and dest branch is `master` - Increments patch version.

    ```text
    v0.1.0 results in v0.1.1
    v1.5.3-pre.2 results in v1.5.4
    ```

- Source branch is prefixed with `feature/` and dest branch is `develop` - Increments minor version.

    ```text
    v0.1.0 results in v0.2.0-pre.1
    v1.5.3-pre.2 results in v1.6.0-pre.1
    ```

- Source branch is prefixed with `major/` and dest branch is `develop` - Increments major version.

    ```text
    v0.1.0 results in v1.0.0-pre.1
    v1.5.3-pre.2 results in v2.0.0-pre.1
    ```

- Source branch is `develop` and dest branch is `master` - Takes the closest tag and finalize it.

    ```text
    v1.5.3-pre.2 results in v1.5.3
    ```

- Source branch is prefixed with `resync/` and dest branch is `develop` - Increments patch version.

    ```text
    v1.5.3-pre.2 results in v1.5.4-pre.1
    ```

## Github Environment Variables

Here are the environment variables we take from Github Actions so far

- `GITHUB_SHA`

## Example usage

### Basic

Uses `auto` bump strategy to calculate the next semantic version.

```yaml
- id: semver-tag
  uses: wakatime/semver-action@vlatest
- name: "Created tag"
  run: echo "tag ${{ steps.semver-tag.outputs.semver_tag }}"
```

### Custom

```yaml
- id: semver-tag
  uses: wakatime/semver-action@vlatest
  with:
    prefix: ""
    prerelease_id: "alpha"
    main_branch_name: "trunk"
    develop_branch_name: "dev"
    debug: "true"
- name: "Created tag"
  run: echo "tag ${{ steps.semver-tag.outputs.semver_tag }}"
```

## Inputs

| parameter           | required | description                                                                      | default     |
| ---                 | ---      | ---                                                                              | ---         |
| bump                |          | Bump strategy for semantic versioning. Can be `auto`, `major`, `minor`, `patch`. | auto        |
| base_version        |          | Version to use as base for the generation, skips version bumps.                  |             |
| prefix              |          | Prefix used to prepend the final version.                                        | v           |
| prerelease_id       |          | Text representing the prerelease identifier.                                    | pre         |
| main_branch_name    |          | The main branch name.                                                            | master      |
| develop_branch_name |          | The develop branch name.                                                         | develop     |
| repo_dir            |          | The repository path.                                                             | current dir |
| debug               |          | Enables debug mode.                                                              | false       |

## Outpus

| parameter     | description                                      |
| ---           | ---                                              |
| semver_tag    | The calculdated semantic version.                |
| is_prerelease | True if calculated tag is prerelease.           |
| previous_tag  | The tag used to calculate next semantic version. |
| ancestor_tag  | The ancestor tag based on specific pattern.      |
