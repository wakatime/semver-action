# Integration Test Cases

1. New repository created with no previous tag and a `major branch` merged into `develop`. The semantic calculated version is `v1.0.0-pre.1` and the previous one is `v0.0.0`.

2. Repository with one merge into `develop`, getting another one to be merged as `feature` into `develop`. The semantic calculated version is `v1.1.0-pre.1` and the previous one is `v1.0.0-pre.1`.

    1. `Base version` is set to `v2.5.17`. The semantic calculated version is `v2.6.0-pre.1` and the previous one is `v1.0.0-pre.1`.

    2. Bump set to `major`. The semantic calculated version is `v2.0.0-pre.1` and the previous one is `v1.0.0-pre.1`.

    3. Bump set to `minor`. The semantic calculated version is `v1.1.0-pre.1` and the previous one is `v1.0.0-pre.1`.

    4. Bump set to `patch`. The semantic calculated version is `v1.0.1-pre.1` and the previous one is `v1.0.0-pre.1`.

3. The `develop` merged into `master` for the first time. The semantic calculated version is `v1.1.0` and the previous one is `v1.1.0-pre.1`.

4. The `Hotfix` merged into `master`. The semantic calculated version is `v1.1.1` and the previous one is `v1.1.0`.
