# Badger

Badger is a Linux eBPF-based PostgreSQL mutation auditor for client processes connecting to Aurora PostgreSQL over TLS.

## Scope

The capture path is independent of the PostgreSQL server. `badgerClan` attaches uprobes to client-side `libpq` and OpenSSL symbols, receives bounded plaintext snapshots through a BPF ring buffer, decodes PostgreSQL protocol messages in Go, and writes an append-only node journal. The namespace is the tenant identifier.

This repository is a prototype. `committed` means that a coherent client-visible command/transaction completion was observed. Network loss, ring-buffer loss, unsupported drivers, process exit, truncation, or incomplete protocol state produce `unknown`; they must not be treated as a regulatory proof of commit.

## Run tests

```sh
go test ./...
```

## Local PostgreSQL client

```sh
cd examples/python-client
uv python install 3.14
uv sync
docker compose up -d
uv run python main.py
```

## Kubernetes security

The Helm chart deliberately uses `hostPID`, `hostNetwork`, host `/proc`, bpffs, debugfs and a privileged container by default. This is required to discover target processes and load uprobes on many managed Kubernetes kernels, but gives the daemonset node-wide process visibility and access to plaintext database parameters. Deploy it only on dedicated trusted nodes, use a restricted service account, encrypt the host journal and use IRSA with least-privilege S3 permissions.

The chart uses namespace as `tenant_id` when Kubernetes metadata is available. Configure an IRSA annotation on the service account and set `s3.bucket`; never put AWS keys in the chart.

## Data path

The local journal is JSONL for inspectability and crash recovery: `events.open.jsonl` is fsynced on every append, then can be atomically renamed to `events.ready.jsonl`. A production deployment should rotate segments, hash-chain records, encrypt the volume, and emit a capture-gap event on ring-buffer loss. Only final `committed` events are selected for Parquet. `unknown` and `failed` remain available for investigation.

## Limitations

The prototype supports PostgreSQL protocol v3 messages and starts with `libpq`/OpenSSL symbols. It does not claim coverage for statically linked clients, non-OpenSSL TLS, COPY, arbitrary binary parameter OIDs, triggers, cascades, server-side functions, or values not present in the client request. Iceberg can reconstruct before/after history only where all relevant committed mutations and row identities were captured.
