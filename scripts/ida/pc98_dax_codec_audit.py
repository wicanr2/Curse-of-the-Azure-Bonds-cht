"""IDA Pro batch audit for the PC-98 DAX loader and block decoder.

Run only against the user's local GAME.EXE database. The commercial binary
and generated IDA database are evidence inputs and must not be committed.
"""

import ida_auto
import ida_bytes
import ida_funcs
import ida_name
import ida_pro
import idautils
import idc


TARGETS = {
    "GETDATABLOCK": 0x17A54,
    "DECODE_PC98_DAX_BLOCK": 0x17DD5,
}


def render(ea):
    size = ida_bytes.get_item_size(ea)
    return (
        f"{ea:08X} bytes={ida_bytes.get_bytes(ea, size).hex()} "
        f"asm={idc.generate_disasm_line(ea, 0)!r}"
    )


ida_auto.auto_wait()
for label, target in TARGETS.items():
    function = ida_funcs.get_func(target)
    print(
        f"FUNCTION label={label} target={target:08X} "
        f"name={ida_name.get_name(function.start_ea) if function else ''!r} "
        f"start={function.start_ea if function else 0:08X} "
        f"end={function.end_ea if function else 0:08X}"
    )
    if not function:
        continue
    for ea in idautils.FuncItems(function.start_ea):
        print(render(ea))

ida_pro.qexit(0)
