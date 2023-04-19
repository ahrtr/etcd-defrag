etcd-defrag
======
etcd-defrag is an easier to use and smarter etcd defragmentation tool. It references the implementation
of `etcdctl defrag` command, but with big refactoring and extra enhancements below,
- run defragmentation only when all members are healthy. Note that it ignores the `NOSPACE` alarm
- run defragmentation on the leader last

etcd-defrag reuses all the existing flags accepted by `etcdctl defrag`, so basically it doesn't break
any existing user experience. Users can just replace `etcdctl defrag [flags]` with `etcd-defrag [flags]`
without compromising any experience.

It adds one more flag `--continue-on-error`. When true, etcd-defrag continues to defragment next endpoint
if current one somehow fails.

# Example
## Example 1: run defragmentation on one endpoint
```
$ ./etcd-defrag --endpoints=http://127.0.0.1:22379
Validating configuration.
Performing health check.
endpoint: http://127.0.0.1:22379, health: true, took: 5.42855ms, error: 
endpoint: http://127.0.0.1:32379, health: true, took: 5.635179ms, error: 
endpoint: http://127.0.0.1:2379, health: true, took: 5.498396ms, error: 
1 endpoints need to be defragmented: [http://127.0.0.1:22379]
Defragmenting endpoint: http://127.0.0.1:22379
Finished defragmenting etcd member[http://127.0.0.1:22379]. took 129.892731ms
The defragmentation is successful.
```

## Example 2: run defragmentation on multiple endpoints
```
$ ./etcd-defrag --endpoints=http://127.0.0.1:22379,http://127.0.0.1:32379
Validating configuration.
Performing health check.
endpoint: http://127.0.0.1:22379, health: true, took: 5.644829ms, error: 
endpoint: http://127.0.0.1:32379, health: true, took: 6.182661ms, error: 
endpoint: http://127.0.0.1:2379, health: true, took: 5.632009ms, error: 
2 endpoints need to be defragmented: [http://127.0.0.1:32379 http://127.0.0.1:22379]
Defragmenting endpoint: http://127.0.0.1:32379
Finished defragmenting etcd member[http://127.0.0.1:32379]. took 176.578231ms
Defragmenting endpoint: http://127.0.0.1:22379
Finished defragmenting etcd member[http://127.0.0.1:22379]. took 131.432431ms
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

## Example 3: run defragmentation on all members in the cluster
```
$ ./etcd-defrag --endpoints http://127.0.0.1:22379 --cluster
Validating configuration.
Performing health check.
endpoint: http://127.0.0.1:22379, health: true, took: 4.405116ms, error: 
endpoint: http://127.0.0.1:32379, health: true, took: 4.622552ms, error: 
endpoint: http://127.0.0.1:2379, health: true, took: 4.929855ms, error: 
3 endpoints need to be defragmented: [http://127.0.0.1:2379 http://127.0.0.1:32379 http://127.0.0.1:22379]
Defragmenting endpoint: http://127.0.0.1:2379
Finished defragmenting etcd member[http://127.0.0.1:2379]. took 128.590789ms
Defragmenting endpoint: http://127.0.0.1:32379
Finished defragmenting etcd member[http://127.0.0.1:32379]. took 128.214737ms
Defragmenting endpoint: http://127.0.0.1:22379
Finished defragmenting etcd member[http://127.0.0.1:22379]. took 125.986093ms
The defragmentation is successful.
```

Only one endpoint is provided, but it still runs defragmentation on all members in the cluster thanks to the flag `--cluster`. 
The leader `http://127.0.0.1:22379` is also placed at the end of the list.

# Contributing
Any contribution is welcome! 

# Plan
- Introduce policy based defragmentation.<br>
  For example, run defragmentation if the db size in use is bigger than 200M and less than 50% of the total db size,
```
db_size_in_use/db_total_size < 0.5 && db_size_in_use > 200M
```

  Note the policy engine should be flexible enough.
- Anything else?
