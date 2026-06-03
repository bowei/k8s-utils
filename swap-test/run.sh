#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SWAP_TEST="$SCRIPT_DIR/swap_test"
TRACER="$SCRIPT_DIR/tracer.sh"
TRACE_FILE="$(mktemp /tmp/swap_trace.XXXXXX.csv)"
TRACER_PID=""
RANDOM_FLAG=""

SIZE_FLAG=""

# Colors
if [[ -t 1 ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    BOLD='\033[1m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    BOLD=''
    NC=''
fi

while [[ $# -gt 0 ]]; do
    case "$1" in
        -r|--random)
            RANDOM_FLAG="--random"
            shift
            ;;
        -s|--size)
            if [[ -n "${2:-}" ]]; then
                SIZE_FLAG="--size $2"
                shift 2
            else
                echo "Error: --size requires an argument" >&2
                exit 1
            fi
            ;;
        *)
            echo "Usage: $0 [-r|--random] [-s|--size <size_mb>]" >&2
            exit 1
            ;;
    esac
done

cleanup() {
    if [[ -n "$TRACER_PID" ]] && kill -0 "$TRACER_PID" 2>/dev/null; then
        kill "$TRACER_PID"
        wait "$TRACER_PID" 2>/dev/null || true
    fi
}
trap cleanup EXIT

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

# Returns the diskstats device name for the active swap.
# Handles both swap partitions and swap files.
detect_swap_device() {
    while read -r filename type _; do
        local dev_node=""
        if [[ "$type" == "partition" ]]; then
            dev_node="$filename"
        elif [[ "$type" == "file" ]]; then
            dev_node=$(df "$filename" 2>/dev/null | awk 'NR==2{print $1}')
        fi

        if [[ -b "$dev_node" ]]; then
            local major_hex minor_hex major minor sys_path
            major_hex=$(stat -c "%t" "$dev_node")
            minor_hex=$(stat -c "%T" "$dev_node")
            major=$((16#$major_hex))
            minor=$((16#$minor_hex))
            if sys_path=$(readlink -f "/sys/dev/block/$major:$minor" 2>/dev/null); then
                basename "$sys_path"
                return
            fi
        fi

        # Fallback
        if [[ -n "$dev_node" ]]; then
            basename "$(readlink -f "$dev_node")"
            return
        fi
    done < <(awk 'NR>1' /proc/swaps)
    echo ""
}

# Snapshot diskstats for a specific device.
# Prints: reads reads_merged read_sectors writes writes_merged write_sectors
snapshot_diskstats() {
    awk -v d="$1" '$3==d{print $4,$5,$6,$8,$9,$10}' /proc/diskstats
}

# Snapshot pswpin pswpout from vmstat.
snapshot_vmstat() {
    awk '/^pswpin/{p=$2}/^pswpout/{o=$2}END{print p+0,o+0}' /proc/vmstat
}

# ---------------------------------------------------------------------------
# Preflight
# ---------------------------------------------------------------------------

if [[ ! -x "$SWAP_TEST" ]]; then
    printf "${RED}ERROR: %s not found — run 'make' first.${NC}\n" "$SWAP_TEST" >&2
    exit 1
fi

SWAP_DEV=$(detect_swap_device)
if [[ -z "$SWAP_DEV" ]]; then
    printf "${RED}ERROR: No active swap device found.${NC}\n" >&2
    exit 1
fi

SWAP_TOTAL=$(awk '/^SwapTotal/{print $2}' /proc/meminfo)
if [[ "$SWAP_TOTAL" -lt $((1024 * 1024)) ]]; then
    printf "${YELLOW}WARNING: Less than 1 GB of swap available (%s kB).${NC}\n" "$SWAP_TOTAL" >&2
fi

printf "${BOLD}%-12s${NC}: %s\n" "Kernel" "$(uname -r)"
printf "${BOLD}%-12s${NC}: %s\n" "Swap device" "$SWAP_DEV"
printf "${BOLD}%-12s${NC}: %s MB\n" "Swap total" "$((SWAP_TOTAL / 1024))"
echo ""

# ---------------------------------------------------------------------------
# Run
# ---------------------------------------------------------------------------

bash "$TRACER" "$SWAP_DEV" 100 > "$TRACE_FILE" &
TRACER_PID=$!

read -r R_BEF RM_BEF RS_BEF W_BEF WM_BEF WS_BEF < <(snapshot_diskstats "$SWAP_DEV")
read -r PIN_BEF POUT_BEF < <(snapshot_vmstat)
T_START_MS=$(date +%s%3N)

TEST_OUTPUT=$("$SWAP_TEST" $RANDOM_FLAG $SIZE_FLAG)
echo "$TEST_OUTPUT"

T_END_MS=$(date +%s%3N)
read -r R_AFT RM_AFT RS_AFT W_AFT WM_AFT WS_AFT < <(snapshot_diskstats "$SWAP_DEV")
read -r PIN_AFT POUT_AFT < <(snapshot_vmstat)

kill "$TRACER_PID"
wait "$TRACER_PID" 2>/dev/null || true
TRACER_PID=""

# ---------------------------------------------------------------------------
# Analysis
# ---------------------------------------------------------------------------

ELAPSED_S=$(awk "BEGIN{printf \"%.3f\", ($T_END_MS - $T_START_MS)/1000}")

awk -v dev="$SWAP_DEV" \
    -v elapsed_s="$ELAPSED_S" \
    -v r_bef="$R_BEF"   -v rm_bef="$RM_BEF"   -v rs_bef="$RS_BEF" \
    -v r_aft="$R_AFT"   -v rm_aft="$RM_AFT"   -v rs_aft="$RS_AFT" \
    -v w_bef="$W_BEF"   -v wm_bef="$WM_BEF"   -v ws_bef="$WS_BEF" \
    -v w_aft="$W_AFT"   -v wm_aft="$WM_AFT"   -v ws_aft="$WS_AFT" \
    -v pin_bef="$PIN_BEF" -v pout_bef="$POUT_BEF" \
    -v pin_aft="$PIN_AFT" -v pout_aft="$POUT_AFT" \
    -v trace="$TRACE_FILE" \
    -v red="$RED" -v green="$GREEN" -v yellow="$YELLOW" -v blue="$BLUE" -v bold="$BOLD" -v nc="$NC" \
'BEGIN {
    # Deltas from before/after snapshots
    dr  = r_aft  - r_bef;   drm = rm_aft - rm_bef;  drs = rs_aft - rs_bef
    dw  = w_aft  - w_bef;   dwm = wm_aft - wm_bef;  dws = ws_aft - ws_bef
    pin  = pin_aft  - pin_bef
    pout = pout_aft - pout_bef

    read_bytes  = drs * 512
    write_bytes = dws * 512

    # Average bandwidth over full test wall time
    avg_r_bw = (elapsed_s > 0) ? read_bytes  / elapsed_s / 1048576 : 0
    avg_w_bw = (elapsed_s > 0) ? write_bytes / elapsed_s / 1048576 : 0

    # Average physical I/O size (post-merge, what the device actually saw)
    avg_r_io = (dr > 0) ? read_bytes  / dr : 0
    avg_w_io = (dw > 0) ? write_bytes / dw : 0

    # Average logical I/O size (pre-merge, what the block layer received)
    avg_r_logical = ((dr + drm) > 0) ? read_bytes  / (dr + drm) : 0
    avg_w_logical = ((dw + dwm) > 0) ? write_bytes / (dw + dwm) : 0

    # Merge ratio: fraction of submitted requests that were absorbed into another
    r_merge_pct = ((dr + drm) > 0) ? drm * 100.0 / (dr + drm) : 0
    w_merge_pct = ((dw + dwm) > 0) ? dwm * 100.0 / (dw + dwm) : 0

    # Peak bandwidth from tracer CSV (per-interval deltas)
    peak_r_bw = 0; peak_w_bw = 0
    prev_t = -1; prev_rs = 0; prev_ws = 0
    while ((getline line < trace) > 0) {
        if (line ~ /^timestamp/) continue   # skip header
        n = split(line, f, ",")
        if (n < 11) continue
        t = f[1]+0; cur_rs = f[6]+0; cur_ws = f[10]+0
        if (prev_t >= 0) {
            dt = (t - prev_t) / 1000.0
            if (dt > 0) {
                rbw = (cur_rs - prev_rs) * 512 / dt / 1048576
                wbw = (cur_ws - prev_ws) * 512 / dt / 1048576
                if (rbw > peak_r_bw) peak_r_bw = rbw
                if (wbw > peak_w_bw) peak_w_bw = wbw
            }
        }
        prev_t = t; prev_rs = cur_rs; prev_ws = cur_ws
    }

    print ""
    print bold "=== Block I/O on " dev " ===" nc
    printf "  %-30s %s %s\n", "", bold "       Reads" nc, bold "      Writes" nc
    printf "  %-30s %12d %12d\n",   "IOs completed (physical):", dr, dw
    printf "  %-30s %12d %12d\n",   "IOs merged (logical):",     drm, dwm
    printf "  %-30s %11.1f%% %11.1f%%\n", "Merge ratio:",        r_merge_pct, w_merge_pct
    printf "  %-30s %11.1f  %11.1f  MB\n","Data transferred:",   read_bytes/1048576, write_bytes/1048576
    print  ""
    printf "  %-30s %11.1f  %11.1f  MB/s\n", "Avg bandwidth (wall time):", avg_r_bw, avg_w_bw
    printf "  %-30s %11.1f  %11.1f  MB/s\n", "Peak bandwidth (100ms win):", peak_r_bw, peak_w_bw
    print  ""
    print  bold "=== I/O Aggregation (how reads are batched to device) ===" nc
    printf "  %-32s %9.1f  KB\n", "Avg physical read size (post-merge):", avg_r_io / 1024
    printf "  %-32s %9.1f  KB\n", "Avg logical read size (pre-merge):",   avg_r_logical / 1024
    if (avg_r_logical > 0)
        printf "  %-32s %9.1fx\n", "Merge amplification factor:", avg_r_io / avg_r_logical
    print  ""
    print  bold "=== VM Swap Pages ===" nc
    printf "  Pages swapped out : %d  (%.1f MB)\n", pout, pout * 4096 / 1048576
    printf "  Pages swapped in  : %d  (%.1f MB)\n", pin,  pin  * 4096 / 1048576
    print  ""
    printf "  Trace saved to: " green "%s" nc "\n", trace
}'

# Prevent cleanup from deleting the trace file
TRACE_FILE=""
