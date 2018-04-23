## Things to do

- check
  - look for summary files in storage
  - parse a timestamp out of filename
- put
  - run command in docker container
    - mount cwd as volume
  - find files according to file glob
    - allow files with `testsuites` or `testsuite` at root
    - add each `testsuite` to a combined parent `testsuites`
    - record overall time as attribute on root `testsuites`
    - record user-specified metadata into `testsuites` attributes, e.g. azure
  - upload summary file to storage
    - ensure that file won't clash on parallel runs
  - print `get` output as on a failed test suite `get` won't run
- get
  - download one or more summary files from storage
  - generate different summaries as specified by pipeline config:
    - summary of last X runs (human-readable)
    - summary of last X runs filtered by metadata (e.g. azure)
    - summary of last X runs filtered by stdout/stderr text (e.g. "502")
    - top X most frequent failing tests over last X runs
    - graphs?

## Notes

- make summary scripts runnable outside of resource, e.g. separate go-gettable CLI
- extract storage implementations into separate library?
- future: support `multi_storage` to aggregate results from multiple teams
- future: benchmark support

## UX

```yaml
resources:
- name: test-runner
  type: test-runner
  source:
    storage:
      driver: git
      uri: git@github.com:concourse/concourse.git
      branch: master
      private_key: ((concourse-repo-private-key))
      # optional
      path_prefix: test-results/
      # OR
      driver: s3
      bucket: my-bucket
      access_key_id: some-key
      secret_access_key: some-secret
      # optional
      path_prefix: test-results/
      # OR
      disabled: true

jobs:
- name: run-tests
  plan:
  - get: ci-repo
  - put: test-runner
    params:
      docker_image: golang:latest
      command: |
        ginkgo -r -p ci-repo/
      results_type: junit
      results_config:
        path: "junit_*.xml"

- name: run-tests-with-private-image
  plan:
  - get: ci-repo
  - get: ci-image
    params:
      save: true
  - put: test-runner
    params:
      docker_image_path: ci-image/image
      command: |
        ginkgo -r -p ci-repo/
      results_type: junit
      results_config:
        path: "junit_*.xml"
```

TODO: Requires a `ginkgo --reporter junit` flag to be useful.
