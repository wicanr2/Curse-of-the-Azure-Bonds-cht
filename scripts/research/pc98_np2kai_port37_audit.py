#!/usr/bin/env python3
"""稽核 instrumented NP2kai 對 exact speaker routine 的 port 37h trace。"""

import json
import re
import sys


EVENT = re.compile(
    r"COAB_PORT37 clock=(?P<clock>\d+) .* "
    r"cs=(?P<cs>[0-9a-fA-F]+) ip=(?P<ip>[0-9a-fA-F]+) "
    r"data=(?P<data>[0-9a-fA-F]+)"
)


def main():
    if len(sys.argv) != 4:
        raise SystemExit(
            "用法：pc98_np2kai_port37_audit.py TRACE.log PERIOD PULSES"
        )
    period = int(sys.argv[2], 0)
    pulses = int(sys.argv[3], 0)
    if period <= 0 or period > 0xFFFF or pulses <= 0 or pulses > 0xFFFF:
        raise SystemExit("PERIOD 與 PULSES 必須介於 1..65535")

    events = []
    with open(sys.argv[1], "r", encoding="utf-8", errors="replace") as source:
        for line in source:
            match = EVENT.search(line)
            if match:
                events.append(
                    {
                        "clock": int(match.group("clock")),
                        "cs": int(match.group("cs"), 16),
                        "ip": int(match.group("ip"), 16),
                        "data": int(match.group("data"), 16),
                    }
                )

    marker = next(
        (
            index
            for index, event in enumerate(events)
            if event["cs"] == 0x1FE0 and event["data"] == 0x05
        ),
        None,
    )
    if marker is None:
        raise SystemExit("找不到 direct-probe data=05 marker")
    exact = events[marker + 1 : marker + 1 + pulses * 2 + 2]
    expected_values = [0x06]
    for _ in range(pulses):
        expected_values.extend((0x06, 0x07))
    expected_values.append(0x07)
    if [event["data"] for event in exact] != expected_values:
        raise SystemExit(
            f"exact OUT values={[event['data'] for event in exact]}，"
            f"預期={expected_values}"
        )
    if any(event["cs"] != 0x1FE0 for event in exact):
        raise SystemExit("exact OUT 不是由 probe CS=1FE0 發出")

    deltas = [
        exact[index]["clock"] - exact[index - 1]["clock"]
        for index in range(1, len(exact))
    ]
    np2_loop_busy = 8 * (period - 1) + 4
    nec_loop_busy = 5 * (period - 1) + 13
    result = {
        "period": period,
        "pulses": pulses,
        "values": [event["data"] for event in exact],
        "ips": [f"{event['ip']:04X}" for event in exact],
        "clocks": [event["clock"] for event in exact],
        "edge_deltas": deltas,
        "np2kai_loop_model": {
            "taken": 8,
            "exit": 4,
            "busy_clocks": np2_loop_busy,
        },
        "nec_v30_execution_model": {
            "taken": 5,
            "exit": 13,
            "busy_clocks": nec_loop_busy,
        },
        "model_busy_clock_difference": np2_loop_busy - nec_loop_busy,
        "classification": (
            "NP2kai-control-flow-exact; "
            "NP2kai-clock-model-not-original-hardware-oracle"
        ),
    }
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
