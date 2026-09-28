# Architecture and Security

`badgerClan` is a privileged node daemon. It discovers application processes, attaches eBPF uprobes, and maps cgroup IDs to Kubernetes pods. The eBPF program only copies bounded user buffers; Go parses protocol frames and applies redaction and correlation in user space.

Preferred hooks are `PQsendQueryParams`, `PQsendQueryPrepared`, `PQexecParams` and result functions. OpenSSL `SSL_write`/`SSL_read` and `_ex` variants provide a generic fallback. Hooks are version and loader dependent, so the collector records the library, symbol, truncation state and source confidence.

An event is finalized as `committed`, `failed`, or `unknown`. A process crash or missing response cannot prove rollback or commit. Before/after values are projections from the captured mutation history; an unavailable value remains null with an unknown confidence.

AWS S3 should be accessed with IRSA. Recommended permissions are `s3:PutObject`, `s3:AbortMultipartUpload`, `s3:ListBucket`, and `s3:HeadObject` scoped to the dataset prefix. Iceberg catalog selection is intentionally left open for AWS Glue Catalog or an S3 Tables integration.
