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
