## TODO

* Write a very minimal git lib
  - Init a new git object with a dir/remote
  - Fetch
  - Get ref of HEAD, ref for any head
* Implement a simple postgres job queue

## Triggers

* Watch a git repo for commits on a particular branch, or for the creation of tags (of a certain form?)
* Same as above but triggered by an HTTP call (github hook) rather than polling
* Expose HTTP API for triggering by remote service
* On a schedule (cron)
* On successful completion of another job (e.g. build triggers test; test triggers deploy)

## Jobs

* Each job runs in its own docker container
* Each job is configured to correspond to one git repo
* Each job is configured to correspond to one docker image
* The repo is mounted in a known (fixed) location on the container
* Config defines a command to run and directory to run it in) ('make', 'fez prod deploy', etc)
* Command status code must be semantically meaningful and determines job success/failure

## Job config

* Job configuration (triggers, output, etc) is defined in a simple format (TOML) and lives in git(?) (instead
  of a traditional DB -- does this make sense?)
* Jobs can be configured to only run one-at-a-time (deploys)
* Jobs can be configured to run on the ref they were triggered by, or on the latest ref on the branchà

## DB

* Job queue lives in postgres
* Job history (limited) lives in postgres too

## Output

* Email results
* Upload artifact to S3
* Expose an HTTP success/failure endpoint (to point watchman at)

## Use cases

* Running unit tests
* Running integration tests (harder, unless we do a fresh DB load on every test run)
* Deploys
* Cron replacement (do a task, alert on failure)

## Questions

* SSH keys/S3 credentials/...

## Libs

* cron
* git
* docker (github.com/fsouza/go-dockerclient?)
* server
