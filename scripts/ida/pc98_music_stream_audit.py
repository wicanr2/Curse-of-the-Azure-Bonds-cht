"""IDA Pro batch audit for MSCDRV track loading and timer-driven playback.

The commercial binary, IDA database, and generated log stay local.  This
script only records code/data-flow evidence needed by the clean remake.
"""

import ida_auto
import ida_bytes
import ida_funcs
import ida_loader
import ida_nalt
import ida_pro
import idautils
import idc


TARGETS = {
    "PUBLIC_7E_HANDLER": 0x10080,
    "TIMER_DISPATCH": 0x10175,
    "SELECT_TRACK_POINTER": 0x1021E,
    "LOAD_TRACK_DESCRIPTORS": 0x10253,
    "STREAM_INTERPRETER": 0x10410,
    "CHANNEL_PARAMETER_HELPER": 0x10CBD,
    "TIMER_INITIALIZE": 0x10CD3,
    "TIMER_STATE_UPDATE": 0x10D3A,
    "TIMER_STOP": 0x10DAA,
    "DIRECT_OPN_WRITE_ADAPTER": 0x1102A,
    "DIRECT_OPN_WRITE": 0x11075,
    "DIRECT_OPN_READ": 0x110AB,
    "SETINTCOND_ADAPTER": 0x11277,
    "SETINTCOND": 0x11288,
}


def render(ea):
    size = ida_bytes.get_item_size(ea)
    data = ida_bytes.get_bytes(ea, size) or b""
    return (
        f"{ea:08X} file={ida_loader.get_fileregion_offset(ea):08X} "
        f"bytes={data.hex()} asm={idc.generate_disasm_line(ea, 0)!r}"
    )


ida_auto.auto_wait()
digest = ida_nalt.retrieve_input_file_sha256()
print(f"INPUT_SHA256 {digest.hex() if digest else ''}")

for label, target in TARGETS.items():
    function = ida_funcs.get_func(target)
    if function is None:
        print(f"FUNCTION label={label} target={target:08X} missing=1")
        continue
    print(
        f"FUNCTION label={label} start={function.start_ea:08X} "
        f"end={function.end_ea:08X}"
    )
    for caller in idautils.CodeRefsTo(function.start_ea, False):
        owner = ida_funcs.get_func(caller)
        owner_start = owner.start_ea if owner else idc.BADADDR
        print(
            f"CALLER target={label} ea={caller:08X} "
            f"function={owner_start:08X} {render(caller)}"
        )
    for ea in idautils.Heads(function.start_ea, function.end_ea):
        if idc.print_insn_mnem(ea).lower() in {
            "call",
            "int",
            "in",
            "out",
            "les",
            "lds",
            "mov",
        }:
            print(render(ea))

print("TRACK_TABLE start=00011600")
for ea in range(0x11600, 0x11618, 2):
    value = ida_bytes.get_word(ea)
    print(
        f"TRACK_POINTER index={(ea - 0x11600) // 2} "
        f"ea={ea:08X} value={value:04X} "
        f"file={ida_loader.get_fileregion_offset(ea):08X}"
    )

ida_pro.qexit(0)
