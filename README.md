etcd-defrag
======
## Table of Contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Overview](#overview)
- [Integration with Kubernetes with a CronJob](#integration-with-kubernetes-with-a-cronjob)
- [Examples](#examples)
  - [Example 1: run defragmentation on one endpoint](#example-1-run-defragmentation-on-one-endpoint)
  - [Example 2: run defragmentation on multiple endpoints](#example-2-run-defragmentation-on-multiple-endpoints)
  - [Example 3: run defragmentation on all members in the cluster](#example-3-run-defragmentation-on-all-members-in-the-cluster)
- [Defragmentation rule](#defragmentation-rule)
- [Container image](#container-image)
- [Contributing](#contributing)
- [Note](#note)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## Overview
etcd-defrag is an easier to use and smarter etcd defragmentation tool. It references the implementation
of `etcdctl defrag` command, but with big refactoring and extra enhancements below,
- check the status of all members, and stop the operation if any member is unhealthy. Note that it ignores the `NOSPACE` alarm
- run defragmentation on the leader last
- support rule based defragmentation

etcd-defrag reuses all the existing flags accepted by `etcdctl defrag`, so basically it doesn't break
any existing user experience, but with additional benefits. Users can just replace `etcdctl defrag [flags]`
with `etcd-defrag [flags]` without compromising any experience.

It adds the following extra flags,
| Flag                         | Description |
|------------------------------|-------------|
| `---compaction`              | whether execute compaction before the defragmentation, defaults to `true` |
| `--continue-on-error`        | whether continue to defragment next endpoint if current one fails, defaults to `true` |
| `--etcd-storage-quota-bytes` | etcd storage quota in bytes (the value passed to etcd instance by flag --quota-backend-bytes), defaults to `2*1024*1024*1024` |
| `--defrag-rule`              | defragmentation rule (etcd-defrag will run defragmentation if the rule is empty or it is evaluated to true), defaults to empty. See more details below. |

See the complete flags below,
```
$ ./etcd-defrag -h
A simple command line tool for etcd defragmentation

Usage:
  etcd-defrag [flags]

Flags:
      --cacert string                  verify certificates of TLS-enabled secure servers using this CA bundle
      --cert string                    identify secure client using this TLS certificate file
      --cluster                        use all endpoints from the cluster member list
      --command-timeout duration       command timeout (excluding dial timeout) (default 30s)
      --compaction                     whether execute compaction before the defragmentation (defaults to true) (default true)
      --continue-on-error              whether continue to defragment next endpoint if current one fails (default true)
      --defrag-rule string             defragmentation rule (etcd-defrag will run defragmentation if the rule is empty or it is evaluated to true)
      --dial-timeout duration          dial timeout for client connections (default 2s)
  -d, --discovery-srv string           domain name to query for SRV records describing cluster endpoints
      --discovery-srv-name string      service name to query when using DNS discovery
      --endpoints strings              comma separated etcd endpoints (default [127.0.0.1:2379])
      --etcd-storage-quota-bytes int   etcd storage quota in bytes (the value passed to etcd instance by flag --quota-backend-bytes) (default 2147483648)
  -h, --help                           help for etcd-defrag
      --insecure-discovery             accept insecure SRV records describing cluster endpoints (default true)
      --insecure-skip-tls-verify       skip server certificate verification (CAUTION: this option should be enabled only for testing purposes)
      --insecure-transport             disable transport security for client connections (default true)
      --keepalive-time duration        keepalive time for client connections (default 2s)
      --keepalive-timeout duration     keepalive timeout for client connections (default 6s)
      --key string                     identify secure client using this TLS key file
      --password string                password for authentication (if this option is used, --user option shouldn't include password)
      --user string                    username[:password] for authentication (prompt if password is not supplied)
      --version                        print the version and exit
```

## Integration with Kubernetes with a CronJob

It is possible to use [the example cronjob in
`./doc/etcd-defrag-cronjob.yaml`](./doc/etcd-defrag-cronjob.yaml) on Kubernetes
environments where the etcd servers are colocated with the control plane nodes.

This example CronJob runs every weekday in the morning, and works by mounting
the `/etc/kubernetes/pki/etcd` folder inside the pod, thereby permitting to
defragment the etcd cluster inside the Kubernetes cluster itself. For more
complex use cases you might to adapt the `--endpoints` and/or the certificates.

The example CronJob is per default configured with
`node-role.kubernetes.io/control-plane` affinity, and with the `hostNetwork:
true` spec, so that the `etcd` server co-located on the apiserver can be
reached directly with `127.0.0.1:2379`.

## Examples
### Example 1: run defragmentation on one endpoint
Command:
```
$ ./etcd-defrag --endpoints=https://127.0.0.1:22379 --cacert ./ca.crt --key ./etcd-defrag.key --cert ./etcd-defrag.crt
```

### Example 2: run defragmentation on multiple endpoints
Command:
```
$ ./etcd-defrag --endpoints=https://127.0.0.1:22379,https://127.0.0.1:32379 --cacert ./ca.crt --key ./etcd-defrag.key --cert ./etcd-defrag.crt
```

### Example 3: run defragmentation on all members in the cluster
Command:
```
$ ./etcd-defrag --endpoints https://127.0.0.1:22379 --cluster --cacert ./ca.crt --key ./etcd-defrag.key --cert ./etcd-defrag.crt
```
Output:
```
Validating configuration.
No defragmentation rule provided
Performing health check.
endpoint: https://127.0.0.1:2379, health: true, took: 4.702492ms, error: 
endpoint: https://127.0.0.1:22379, health: true, took: 5.017075ms, error: 
endpoint: https://127.0.0.1:32379, health: true, took: 4.747068ms, error: 
Getting members status
endpoint: https://127.0.0.1:2379, dbSize: 172032, dbSizeInUse: 126976, memberId: 8211f1d0f64f3269, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10425
endpoint: https://127.0.0.1:22379, dbSize: 122880, dbSizeInUse: 122880, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10425
endpoint: https://127.0.0.1:32379, dbSize: 122880, dbSizeInUse: 122880, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10425
Running compaction until revision: 10365 ... successful
3 endpoint(s) need to be defragmented: [https://127.0.0.1:22379 https://127.0.0.1:32379 https://127.0.0.1:2379]
[Before defragmentation] endpoint: https://127.0.0.1:22379, dbSize: 126976, dbSizeInUse: 90112, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10426
Defragmenting endpoint "https://127.0.0.1:22379"
Finished defragmenting etcd endpoint "https://127.0.0.1:22379". took 224.151378ms
[Post defragmentation] endpoint: https://127.0.0.1:22379, dbSize: 90112, dbSizeInUse: 81920, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10426
[Before defragmentation] endpoint: https://127.0.0.1:32379, dbSize: 126976, dbSizeInUse: 90112, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10426
Defragmenting endpoint "https://127.0.0.1:32379"
Finished defragmenting etcd endpoint "https://127.0.0.1:32379". took 139.138035ms
[Post defragmentation] endpoint: https://127.0.0.1:32379, dbSize: 90112, dbSizeInUse: 81920, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10426
[Before defragmentation] endpoint: https://127.0.0.1:2379, dbSize: 172032, dbSizeInUse: 94208, memberId: 8211f1d0f64f3269, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10426
Defragmenting endpoint "https://127.0.0.1:2379"
Finished defragmenting etcd endpoint "https://127.0.0.1:2379". took 135.171807ms
[Post defragmentation] endpoint: https://127.0.0.1:2379, dbSize: 90112, dbSizeInUse: 81920, memberId: 8211f1d0f64f3269, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10426
The defragmentation is successful.
```

Only one endpoint is provided, but it still runs defragmentation on all members in the cluster thanks to the flag `--cluster`.
Note that the endpoint `https://127.0.0.1:2379` is the leader, so it's placed at the end of the list,
```
3 endpoint(s) need to be defragmented: [https://127.0.0.1:22379 https://127.0.0.1:32379 https://127.0.0.1:2379]
```
```
$ etcdctl endpoint status -w table --cluster
+-------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|        ENDPOINT         |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+-------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|  https://127.0.0.1:2379 | 8211f1d0f64f3269 |   3.5.8 |   25 kB |      true |      false |        10 |        164 |                164 |        |
| https://127.0.0.1:22379 | 91bc3c398fb3c146 |   3.5.8 |   25 kB |     false |      false |        10 |        164 |                164 |        |
| https://127.0.0.1:32379 | fd422379fda50e48 |   3.5.8 |   25 kB |     false |      false |        10 |        164 |                164 |        |
+-------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
```
## Defragmentation rule
Defragmentation is an expensive operation, so it should be executed as infrequent as possible. On the other hand, 
it's also necessary to make sure any etcd member will not run out of the storage quota. It's exactly the reason 
why the defragmentation rule is introduced, it can skip unnecessary expensive defragmentation, and also keep
each member safe.

Users can configure a defragmentation rule using the flag `--defrag-rule`. The rule must be a boolean expression,
which means its evaluation result should be a boolean value. **It supports arithmetic (e.g. `+` `-` `*` `/` `%`) and logic
(e.g. `==` `!=` `<` `>` `<=` `>=` `&&` `||` `!`) operators supported by golang. Parenthesis `()` can be used to control precedence**.

Currently, `etcd-defrag` supports three variables below,
| Variable name   | Description |
|---------------  |-------------|
| `dbSize`        | total size of the etcd database |
| `dbSizeInUse`   | total size in use of the etcd database |
| `dbSizeFree`    | total size not in use of the etcd database, defined as dbSize - dbSizeInUse|
| `dbQuota`       | etcd storage quota in bytes (the value passed to etcd instance by flag --quota-backend-bytes)|
| `dbQuotaUsage`  | total usage of the etcd storage quota, defined as dbSize/dbQuota |

For example, if you want to run defragmentation if the total db size is greater than 80%
of the quota **OR** there is at least 200MiB free space, the defragmentation rule is `dbSize > dbQuota*80/100 || dbSize - dbSizeInUse > 200*1024*1024`.
The complete command is below,
```
$ ./etcd-defrag --endpoints http://127.0.0.1:22379 --cluster --defrag-rule="dbSize > dbQuota*80/100 || dbSize - dbSizeInUse > 200*1024*1024"
```
Or,
```
$ ./etcd-defrag --endpoints http://127.0.0.1:22379 --cluster --defrag-rule="dbQuotaUsage > 0.8 || dbSizeFree > 200*1024*1024"
```

Output:
```
Validating configuration.
Validating the defragmentation rule: dbSize > dbQuota*80/100 || dbSize - dbSizeInUse > 200*1024*1024 ... valid
Performing health check.
endpoint: http://127.0.0.1:2379, health: true, took: 6.993264ms, error: 
endpoint: http://127.0.0.1:32379, health: true, took: 7.483368ms, error: 
endpoint: http://127.0.0.1:22379, health: true, took: 49.441931ms, error: 
Getting members status
endpoint: http://127.0.0.1:2379, dbSize: 131072, dbSizeInUse: 131072, memberId: 8211f1d0f64f3269, leader: 8211f1d0f64f3269, revision: 10964, term: 2, index: 11028
endpoint: http://127.0.0.1:22379, dbSize: 131072, dbSizeInUse: 131072, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 10964, term: 2, index: 11028
endpoint: http://127.0.0.1:32379, dbSize: 131072, dbSizeInUse: 131072, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269, revision: 10964, term: 2, index: 11028
Running compaction until revision: 10964 ... successful
3 endpoint(s) need to be defragmented: [http://127.0.0.1:22379 http://127.0.0.1:32379 http://127.0.0.1:2379]
[Before defragmentation] endpoint: http://127.0.0.1:22379, dbSize: 139264, dbSizeInUse: 90112, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 10964, term: 2, index: 11029
Evaluation result is false, so skipping endpoint: http://127.0.0.1:22379
[Before defragmentation] endpoint: http://127.0.0.1:32379, dbSize: 139264, dbSizeInUse: 139264, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269, revision: 10964, term: 2, index: 11029
Evaluation result is false, so skipping endpoint: http://127.0.0.1:32379
[Before defragmentation] endpoint: http://127.0.0.1:2379, dbSize: 139264, dbSizeInUse: 90112, memberId: 8211f1d0f64f3269, leader: 8211f1d0f64f3269, revision: 10964, term: 2, index: 11029
Evaluation result is false, so skipping endpoint: http://127.0.0.1:2379
The defragmentation is successful.
```

If you want to run defragmentation when both conditions are true, namely the total db size is greater than 80%
of the quota **AND** there is at least 200MiB free space, then run command below,
```
$ ./etcd-defrag --endpoints http://127.0.0.1:22379 --cluster --defrag-rule="dbSize > dbQuota*80/100 && dbSize - dbSizeInUse > 200*1024*1024"
```

## Container image
Container images are released automatically using GitHub actions and [`ko-build/ko`](https://github.com/ko-build/ko).
They can be used as follows:

```bash
$ docker pull ghcr.io/ahrtr/etcd-defrag:latest
```

Alternatively, you can build your own container images with:

```bash
$ DOCKER_BUILDKIT=1 docker build -t "etcd-defrag:${VERSION}" -f Dockerfile .
```

If you need an image for another `GOARCH` (e.g. `arm64`, `ppc64le` or `s390x`) other than `amd64`, use a command something like below,
```bash
$ DOCKER_BUILDKIT=1 docker build --build-arg ARCH=${ARCH} -t "etcd-defrag:${VERSION}" -f Dockerfile .
```

## Contributing
Any contribution is welcome!

## Note
Please ensure running etcd on a version >= 3.5.6, and read [Two possible data inconsistency issues in etcd](https://groups.google.com/g/etcd-dev/c/8S7u6NqW6C4) to get more details.
