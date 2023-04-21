etcd-defrag
======
## Table of Contents

- **[Overview](#overview)**
- **[Examples](#examples)**
  - [Example 1: run defragmentation on one endpoint](#example-1-run-defragmentation-on-one-endpoint)
  - [Example 2: run defragmentation on multiple endpoints](#example-2-run-defragmentation-on-multiple-endpoints)
  - [Example 3: run defragmentation on all members in the cluster](#example-3-run-defragmentation-on-all-members-in-the-cluster)
- **[Defragmentation rules](#defragmentation-rules)**
- **[Contributing](#contributing)**
- **[Note](#note)**

## Overview
etcd-defrag is an easier to use and smarter etcd defragmentation tool. It references the implementation
of `etcdctl defrag` command, but with big refactoring and extra enhancements below,
- check the status of all members, and stop the operation if any member is unhealthy. Note that it ignores the `NOSPACE` alarm
- run defragmentation on the leader last
- support rules based defragmentation

etcd-defrag reuses all the existing flags accepted by `etcdctl defrag`, so basically it doesn't break
any existing user experience, but with additional benefits. Users can just replace `etcdctl defrag [flags]`
with `etcd-defrag [flags]` without compromising any experience.

It adds the following extra flags,
| Flag                         | Description |
|------------------------------|-------------|
| `--continue-on-error`        | whether continue to defragment next endpoint if current one fails, defaults to `true` |
| `--etcd-storage-quota-bytes` | etcd storage quota in bytes (the value passed to etcd instance by flag --quota-backend-bytes), defaults to `2*1024*1024*1024` |
| `--defrag-rules`             | comma separated rules (etcd-defrag will run defragmentation if the rule is empty or any rule is evaluated to true), defaults to empty. See more details below. |


## Examples
### Example 1: run defragmentation on one endpoint
Command:
```
$ ./etcd-defrag --endpoints=http://127.0.0.1:22379
```
Output:
```
Validating configuration.
There are 0 defragmentation rules: []
Performing health check.
endpoint: http://127.0.0.1:32379, health: true, took: 3.947936ms, error: 
endpoint: http://127.0.0.1:22379, health: true, took: 3.722515ms, error: 
endpoint: http://127.0.0.1:2379, health: true, took: 3.781799ms, error: 
Getting members status
endpoint: http://127.0.0.1:22379, dbSize: 24576, dbSizeInUse: 16384, memberId: 91bc3c398fb3c146, leader: fd422379fda50e48
1 endpoints need to be defragmented: [http://127.0.0.1:22379]
Defragmenting endpoint: http://127.0.0.1:22379
Finished defragmenting etcd member[http://127.0.0.1:22379]. took 124.183091ms
The defragmentation is successful.
```

### Example 2: run defragmentation on multiple endpoints
Command:
```
$ ./etcd-defrag --endpoints=http://127.0.0.1:22379,http://127.0.0.1:32379
```
Output:
```
Validating configuration.
There are 0 defragmentation rules: []
Performing health check.
endpoint: http://127.0.0.1:2379, health: true, took: 5.285641ms, error: 
endpoint: http://127.0.0.1:32379, health: true, took: 5.349508ms, error: 
endpoint: http://127.0.0.1:22379, health: true, took: 5.193551ms, error: 
Getting members status
endpoint: http://127.0.0.1:22379, dbSize: 24576, dbSizeInUse: 24576, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269
endpoint: http://127.0.0.1:32379, dbSize: 24576, dbSizeInUse: 24576, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269
2 endpoints need to be defragmented: [http://127.0.0.1:22379 http://127.0.0.1:32379]
Defragmenting endpoint: http://127.0.0.1:22379
Finished defragmenting etcd member[http://127.0.0.1:22379]. took 204.570105ms
Defragmenting endpoint: http://127.0.0.1:32379
Finished defragmenting etcd member[http://127.0.0.1:32379]. took 122.612044ms
The defragmentation is successful.
```

Note that the endpoint `http://127.0.0.1:22379` is the leader, so it's placed at the end of the list,
```
$ etcdctl endpoint status -w table --cluster
+------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|        ENDPOINT        |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|  http://127.0.0.1:2379 | 8211f1d0f64f3269 |   3.5.8 |   25 kB |     false |      false |         6 |         37 |                 37 |        |
| http://127.0.0.1:22379 | 91bc3c398fb3c146 |   3.5.8 |   25 kB |      true |      false |         6 |         37 |                 37 |        |
| http://127.0.0.1:32379 | fd422379fda50e48 |   3.5.8 |   25 kB |     false |      false |         6 |         37 |                 37 |        |
+------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
```

### Example 3: run defragmentation on all members in the cluster
Command:
```
$ ./etcd-defrag --endpoints http://127.0.0.1:22379 --cluster
```
Output:
```
Validating configuration.
There are 0 defragmentation rules: []
Performing health check.
endpoint: http://127.0.0.1:2379, health: true, took: 3.781485ms, error: 
endpoint: http://127.0.0.1:32379, health: true, took: 3.989011ms, error: 
endpoint: http://127.0.0.1:22379, health: true, took: 5.09677ms, error: 
Getting members status
endpoint: http://127.0.0.1:2379, dbSize: 24576, dbSizeInUse: 24576, memberId: 8211f1d0f64f3269, leader: 8211f1d0f64f3269
endpoint: http://127.0.0.1:22379, dbSize: 24576, dbSizeInUse: 16384, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269
endpoint: http://127.0.0.1:32379, dbSize: 24576, dbSizeInUse: 16384, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269
3 endpoints need to be defragmented: [http://127.0.0.1:22379 http://127.0.0.1:32379 http://127.0.0.1:2379]
Defragmenting endpoint: http://127.0.0.1:22379
Finished defragmenting etcd member[http://127.0.0.1:22379]. took 124.132735ms
Defragmenting endpoint: http://127.0.0.1:32379
Finished defragmenting etcd member[http://127.0.0.1:32379]. took 122.426241ms
Defragmenting endpoint: http://127.0.0.1:2379
Finished defragmenting etcd member[http://127.0.0.1:2379]. took 122.99698ms
The defragmentation is successful.
```

Only one endpoint is provided, but it still runs defragmentation on all members in the cluster thanks to the flag `--cluster`.
Note that the endpoint `http://127.0.0.1:2379` is the leader, so it's placed at the end of the list,
```
3 endpoints need to be defragmented: [http://127.0.0.1:22379 http://127.0.0.1:32379 http://127.0.0.1:2379]
```
```
$ etcdctl endpoint status -w table --cluster
+------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|        ENDPOINT        |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|  http://127.0.0.1:2379 | 8211f1d0f64f3269 |   3.5.8 |   25 kB |      true |      false |        10 |        164 |                164 |        |
| http://127.0.0.1:22379 | 91bc3c398fb3c146 |   3.5.8 |   25 kB |     false |      false |        10 |        164 |                164 |        |
| http://127.0.0.1:32379 | fd422379fda50e48 |   3.5.8 |   25 kB |     false |      false |        10 |        164 |                164 |        |
+------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
```

## Defragmentation rules
Users can configure comma separated defragmentation rules using the flag `--defrag-rules`. Each rule must be a boolean expression,
which means its evaluation result should be a boolean value. **It supports arithmetic (e.g. `+` `-` `*` `/` `%`) and logic
(e.g. `==` `!=` `<` `>` `<=` `>=` `&&` `||` `!`) operators supported by golang. Parenthesis `()` can be used to control precedence**.

Currently, `etcd-defrag` supports three variables below,
| Variable name   | Description |
|---------------  |-------------|
| `dbSize`        | total size of the etcd database |
| `dbSizeInUse`   | total size in use of the etcd database |
| `dbQuota`       | etcd storage quota in bytes (the value passed to etcd instance by flag --quota-backend-bytes)|

For example, if you want to run defragmentation if the total db size is greater than 80%
of the quota **OR** there is at least 200MiB free space, the defragmentation rules are `dbSize > dbQuota*80/100, dbSize - dbSizeInUse > 200*1024*1024`.
The complete command is below,
```
$ ./etcd-defrag --endpoints http://127.0.0.1:22379 --cluster --defrag-rules="dbSize > dbQuota*80/100, dbSize - dbSizeInUse > 200*1024*1024"
```
Output:
```
Validating configuration.
There are 2 defragmentation rules: [dbSize*80/100 > dbQuota  dbSize - dbSizeInUse > 200*1024*1024]
Rule "dbSize*80/100 > dbQuota" is valid.
Rule " dbSize - dbSizeInUse > 200*1024*1024" is valid.
Performing health check.
endpoint: http://127.0.0.1:2379, health: true, took: 3.176305ms, error: 
endpoint: http://127.0.0.1:22379, health: true, took: 2.978166ms, error: 
endpoint: http://127.0.0.1:32379, health: true, took: 4.379929ms, error: 
Getting members status
endpoint: http://127.0.0.1:2379, dbSize: 24576, dbSizeInUse: 16384, memberId: 8211f1d0f64f3269, leader: 8211f1d0f64f3269
endpoint: http://127.0.0.1:22379, dbSize: 24576, dbSizeInUse: 16384, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269
endpoint: http://127.0.0.1:32379, dbSize: 24576, dbSizeInUse: 16384, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269
3 endpoints need to be defragmented: [http://127.0.0.1:22379 http://127.0.0.1:32379 http://127.0.0.1:2379]
Defragmenting endpoint: http://127.0.0.1:22379
Evaluation result is false, so skipping endpoint: http://127.0.0.1:22379
Defragmenting endpoint: http://127.0.0.1:32379
Evaluation result is false, so skipping endpoint: http://127.0.0.1:32379
Defragmenting endpoint: http://127.0.0.1:2379
Evaluation result is false, so skipping endpoint: http://127.0.0.1:2379
The defragmentation is successful.
```

The above command is equivalent to command below,
```
$ ./etcd-defrag --endpoints http://127.0.0.1:22379 --cluster --defrag-rules="dbSize > dbQuota*80/100 || dbSize - dbSizeInUse > 200*1024*1024"
```

If you want to run defragmentation when both conditions are true, namely the total db size is greater than 80%
of the quota **AND** there is at least 200MiB free space, then run command below,
```
$ ./etcd-defrag --endpoints http://127.0.0.1:22379 --cluster --defrag-rules="dbSize > dbQuota*80/100 && dbSize - dbSizeInUse > 200*1024*1024"
```

## Contributing
Any contribution is welcome!

## Note
Please ensure running etcd on a version > 3.5.4, and read [Two possible data inconsistency issues in etcd](https://groups.google.com/g/etcd-dev/c/8S7u6NqW6C4) to get more details.
