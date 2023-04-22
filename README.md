etcd-defrag
======
## Table of Contents

- **[Overview](#overview)**
- **[Examples](#examples)**
  - [Example 1: run defragmentation on one endpoint](#example-1-run-defragmentation-on-one-endpoint)
  - [Example 2: run defragmentation on multiple endpoints](#example-2-run-defragmentation-on-multiple-endpoints)
  - [Example 3: run defragmentation on all members in the cluster](#example-3-run-defragmentation-on-all-members-in-the-cluster)
- **[Defragmentation rule](#defragmentation-rule)**
- **[Contributing](#contributing)**
- **[Note](#note)**

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


## Examples
### Example 1: run defragmentation on one endpoint
Command:
```
$ ./etcd-defrag --endpoints=http://127.0.0.1:22379
```
Output:
```
Validating configuration.
No defragmentation rule provided
Performing health check.
endpoint: http://127.0.0.1:22379, health: true, took: 7.227642ms, error: 
endpoint: http://127.0.0.1:2379, health: true, took: 13.255694ms, error: 
endpoint: http://127.0.0.1:32379, health: true, took: 6.666809ms, error: 
Getting members status
endpoint: http://127.0.0.1:22379, dbSize: 167936, dbSizeInUse: 167936, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 9779, term: 2, index: 9831
Running compaction until revision: 9779 ... successful
1 endpoint(s) need to be defragmented: [http://127.0.0.1:22379]
[Before defragmentation] endpoint: http://127.0.0.1:22379, dbSize: 167936, dbSizeInUse: 94208, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 9779, term: 2, index: 9832
Defragmenting endpoint "http://127.0.0.1:22379"
Finished defragmenting etcd endpoint "http://127.0.0.1:22379". took 161.063637ms
[Post defragmentation] endpoint: http://127.0.0.1:22379, dbSize: 90112, dbSizeInUse: 81920, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 9779, term: 2, index: 9832
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
No defragmentation rule provided
Performing health check.
endpoint: http://127.0.0.1:2379, health: true, took: 6.368905ms, error: 
endpoint: http://127.0.0.1:22379, health: true, took: 6.497803ms, error: 
endpoint: http://127.0.0.1:32379, health: true, took: 6.745877ms, error: 
Getting members status
endpoint: http://127.0.0.1:22379, dbSize: 106496, dbSizeInUse: 106496, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 9907, term: 2, index: 9963
endpoint: http://127.0.0.1:32379, dbSize: 167936, dbSizeInUse: 106496, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269, revision: 9907, term: 2, index: 9963
Running compaction until revision: 9907 ... successful
2 endpoint(s) need to be defragmented: [http://127.0.0.1:22379 http://127.0.0.1:32379]
[Before defragmentation] endpoint: http://127.0.0.1:22379, dbSize: 110592, dbSizeInUse: 94208, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 9907, term: 2, index: 9964
Defragmenting endpoint "http://127.0.0.1:22379"
Finished defragmenting etcd endpoint "http://127.0.0.1:22379". took 171.412229ms
[Post defragmentation] endpoint: http://127.0.0.1:22379, dbSize: 90112, dbSizeInUse: 81920, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 9907, term: 2, index: 9964
[Before defragmentation] endpoint: http://127.0.0.1:32379, dbSize: 167936, dbSizeInUse: 94208, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269, revision: 9907, term: 2, index: 9964
Defragmenting endpoint "http://127.0.0.1:32379"
Finished defragmenting etcd endpoint "http://127.0.0.1:32379". took 132.445712ms
[Post defragmentation] endpoint: http://127.0.0.1:32379, dbSize: 90112, dbSizeInUse: 81920, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269, revision: 9907, term: 2, index: 9964
The defragmentation is successful.
```

### Example 3: run defragmentation on all members in the cluster
Command:
```
$ ./etcd-defrag --endpoints http://127.0.0.1:22379 --cluster
```
Output:
```
Validating configuration.
No defragmentation rule provided
Performing health check.
endpoint: http://127.0.0.1:2379, health: true, took: 4.702492ms, error: 
endpoint: http://127.0.0.1:22379, health: true, took: 5.017075ms, error: 
endpoint: http://127.0.0.1:32379, health: true, took: 4.747068ms, error: 
Getting members status
endpoint: http://127.0.0.1:2379, dbSize: 172032, dbSizeInUse: 126976, memberId: 8211f1d0f64f3269, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10425
endpoint: http://127.0.0.1:22379, dbSize: 122880, dbSizeInUse: 122880, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10425
endpoint: http://127.0.0.1:32379, dbSize: 122880, dbSizeInUse: 122880, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10425
Running compaction until revision: 10365 ... successful
3 endpoint(s) need to be defragmented: [http://127.0.0.1:22379 http://127.0.0.1:32379 http://127.0.0.1:2379]
[Before defragmentation] endpoint: http://127.0.0.1:22379, dbSize: 126976, dbSizeInUse: 90112, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10426
Defragmenting endpoint "http://127.0.0.1:22379"
Finished defragmenting etcd endpoint "http://127.0.0.1:22379". took 224.151378ms
[Post defragmentation] endpoint: http://127.0.0.1:22379, dbSize: 90112, dbSizeInUse: 81920, memberId: 91bc3c398fb3c146, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10426
[Before defragmentation] endpoint: http://127.0.0.1:32379, dbSize: 126976, dbSizeInUse: 90112, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10426
Defragmenting endpoint "http://127.0.0.1:32379"
Finished defragmenting etcd endpoint "http://127.0.0.1:32379". took 139.138035ms
[Post defragmentation] endpoint: http://127.0.0.1:32379, dbSize: 90112, dbSizeInUse: 81920, memberId: fd422379fda50e48, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10426
[Before defragmentation] endpoint: http://127.0.0.1:2379, dbSize: 172032, dbSizeInUse: 94208, memberId: 8211f1d0f64f3269, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10426
Defragmenting endpoint "http://127.0.0.1:2379"
Finished defragmenting etcd endpoint "http://127.0.0.1:2379". took 135.171807ms
[Post defragmentation] endpoint: http://127.0.0.1:2379, dbSize: 90112, dbSizeInUse: 81920, memberId: 8211f1d0f64f3269, leader: 8211f1d0f64f3269, revision: 10365, term: 2, index: 10426
The defragmentation is successful.
```

Only one endpoint is provided, but it still runs defragmentation on all members in the cluster thanks to the flag `--cluster`.
Note that the endpoint `http://127.0.0.1:2379` is the leader, so it's placed at the end of the list,
```
3 endpoint(s) need to be defragmented: [http://127.0.0.1:22379 http://127.0.0.1:32379 http://127.0.0.1:2379]
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

## Defragmentation rule
Users can configure a defragmentation rule using the flag `--defrag-rule`. The rule must be a boolean expression,
which means its evaluation result should be a boolean value. **It supports arithmetic (e.g. `+` `-` `*` `/` `%`) and logic
(e.g. `==` `!=` `<` `>` `<=` `>=` `&&` `||` `!`) operators supported by golang. Parenthesis `()` can be used to control precedence**.

Currently, `etcd-defrag` supports three variables below,
| Variable name   | Description |
|---------------  |-------------|
| `dbSize`        | total size of the etcd database |
| `dbSizeInUse`   | total size in use of the etcd database |
| `dbQuota`       | etcd storage quota in bytes (the value passed to etcd instance by flag --quota-backend-bytes)|

For example, if you want to run defragmentation if the total db size is greater than 80%
of the quota **OR** there is at least 200MiB free space, the defragmentation rule is `dbSize > dbQuota*80/100 || dbSize - dbSizeInUse > 200*1024*1024`.
The complete command is below,
```
$ ./etcd-defrag --endpoints http://127.0.0.1:22379 --cluster --defrag-rule="dbSize > dbQuota*80/100 || dbSize - dbSizeInUse > 200*1024*1024"
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

## Contributing
Any contribution is welcome!

## Note
Please ensure running etcd on a version >= 3.5.6, and read [Two possible data inconsistency issues in etcd](https://groups.google.com/g/etcd-dev/c/8S7u6NqW6C4) to get more details.
