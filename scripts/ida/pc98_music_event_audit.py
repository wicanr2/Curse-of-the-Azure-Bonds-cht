"""IDA Pro batch evidence for MSCDRV sequence side effects.

Run only against the user's local MSCDRV.EXE database.  The commercial
binary, database, and generated log stay local; this script is safe to keep
because it contains addresses and extraction logic, not game data.
"""

import ida_auto
import ida_bytes
import ida_funcs
import ida_loader
import ida_nalt
import ida_pro
import ida_segment
import idautils
import idc


FUNCTIONS = {
    "STREAM_INTERPRETER": 0x10410,
    "PSG_REGISTER_ADAPTER": 0x10CBD,
    "DIRECT_OPN_WRITE_ADAPTER": 0x11189,
    "TEMPO_REGISTER_ADAPTER": 0x111E2,
    "SETVOLUME_ADAPTER": 0x1128F,
    "SETPARABLOCK_ADAPTER": 0x112A2,
    "SETPARABLOCK_ADDRESS": 0x112B0,
}

DATA_RANGES = {
    "FM_FNUMBER_TABLE_DS_0210": (0x114E0, 0x114F8),
    "PSG_PERIOD_TABLE_DS_0228": (0x114F8, 0x11586),
    "TRACK_POINTER_TABLE_DS_0330": (0x11600, 0x11618),
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

for index in range(ida_segment.get_segm_qty()):
    segment = ida_segment.getnseg(index)
    print(
        f"SEGMENT name={ida_segment.get_segm_name(segment)} "
        f"start={segment.start_ea:08X} end={segment.end_ea:08X}"
    )

for label, target in FUNCTIONS.items():
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
        print(render(ea))

for label, (start, end) in DATA_RANGES.items():
    data = ida_bytes.get_bytes(start, end - start) or b""
    print(
        f"DATA label={label} start={start:08X} end={end:08X} "
        f"file={ida_loader.get_fileregion_offset(start):08X} bytes={data.hex()}"
    )

ida_pro.qexit(0)
