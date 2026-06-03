#!/usr/bin/env bash
# Polls /proc/vmstat and /proc/diskstats for a specific block device.
# Usage: tracer.sh <device_name> [interval_ms]
# Outputs raw cumulative counter CSV; run.sh computes derived stats.
#
# /proc/diskstats field layout (after major/minor/name):
#   $4  reads_completed   $5  reads_merged    $6  sectors_read    $7  ms_reading
#   $8  writes_completed  $9  writes_merged   $10 sectors_written  $11 ms_writing

DEVICE=${1:?Usage: tracer.sh <device_name> [interval_ms]}
INTERVAL_MS=${2:-100}
SLEEP_S=$(awk "BEGIN{printf \"%.3f\", $INTERVAL_MS/1000}")

echo "timestamp_ms,pswpin,pswpout,reads,reads_merged,read_sectors,ms_reading,writes,writes_merged,write_sectors,ms_writing"

cleanup() { exit 0; }
trap cleanup SIGTERM SIGINT

START_MS=$(date +%s%3N)

while true; do
    NOW_MS=$(date +%s%3N)
    read -r PIN POUT < <(awk '/^pswpin/{p=$2}/^pswpout/{o=$2}END{print p+0,o+0}' /proc/vmstat)
    read -r R RM RS MR W WM WS MW < <(awk -v d="$DEVICE" '$3==d{print $4,$5,$6,$7,$8,$9,$10,$11}' /proc/diskstats)
    echo "$((NOW_MS-START_MS)),${PIN:-0},${POUT:-0},${R:-0},${RM:-0},${RS:-0},${MR:-0},${W:-0},${WM:-0},${WS:-0},${MW:-0}"
    sleep "$SLEEP_S"
done
