//go:build ignore
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

#define MAX_PAYLOAD 8192
struct event { __u64 pid_tgid, cgroup_id, timestamp_ns; __u32 length; __u8 direction; __u8 source; char payload[MAX_PAYLOAD]; };
struct { __uint(type, BPF_MAP_TYPE_RINGBUF); __uint(max_entries, 16 << 20); } events SEC(".maps");

static int capture(const void *buf, __u64 size, __u8 direction, __u8 source) {
    struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0); if (!e) return 0;
    e->pid_tgid = bpf_get_current_pid_tgid(); e->cgroup_id = bpf_get_current_cgroup_id(); e->timestamp_ns = bpf_ktime_get_ns();
    e->length = size > MAX_PAYLOAD ? MAX_PAYLOAD : size; e->direction = direction; e->source = source;
    if (bpf_probe_read_user(e->payload, e->length, buf) < 0) { bpf_ringbuf_discard(e, 0); return 0; }
    bpf_ringbuf_submit(e, 0); return 0;
}

SEC("uprobe/SSL_write") int ssl_write(struct pt_regs *ctx) { return capture((void *)PT_REGS_PARM2(ctx), PT_REGS_PARM3(ctx), 1, 2); }
SEC("uprobe/SSL_read") int ssl_read(struct pt_regs *ctx) { return capture((void *)PT_REGS_PARM2(ctx), PT_REGS_PARM3(ctx), 0, 2); }
SEC("uprobe/libpq") int libpq_query(struct pt_regs *ctx) { return capture((void *)PT_REGS_PARM2(ctx), 4096, 1, 1); }
char LICENSE[] SEC("license") = "GPL";
