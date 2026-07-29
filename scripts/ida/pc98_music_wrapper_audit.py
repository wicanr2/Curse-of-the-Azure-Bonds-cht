"""IDA Pro batch audit for the PC-98 music wrapper and IVT trampoline.

Run only against the user's local GAME.EXE database. The commercial binary
and generated IDA database are evidence inputs and must not be committed.
"""

import ida_auto
import ida_bytes
import ida_funcs
import ida_loader
import ida_name
import ida_pro
import idautils
import idc


TARGETS = {
    "INITSOUND": 0x18A3D,
    "MSCPLAY": 0x18A44,
    "MSCSTOP": 0x18A8E,
    "BGMPLAY": 0x18AA7,
    "IVT_FAR_TRAMPOLINE": 0x18BDB,
}


def render(ea):
    size = ida_bytes.get_item_size(ea)
    data = ida_bytes.get_bytes(ea, size) or b""
    return (
        f"{ea:08X} file={ida_loader.get_fileregion_offset(ea):08X} "
        f"bytes={data.hex()} "
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
    for ref in idautils.XrefsTo(function.start_ea, 0):
        print(
            f"XREF target={label} from={ref.frm:08X} "
            f"type={ref.type} iscode={ref.iscode}"
        )

for needle_label, needle in (("VECTOR_7E", 0x7E), ("INT_D2", 0xD2)):
    print(f"IMMEDIATE_SCAN label={needle_label} value={needle:02X}")
    for segment_ea in idautils.Segments():
        segment_end = idc.get_segm_end(segment_ea)
        for ea in idautils.Heads(segment_ea, segment_end):
            if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
                continue
            if any(idc.get_operand_value(ea, index) == needle for index in range(2)):
                print(render(ea))

print("TRAMPOLINE_RETURN")
for ea in idautils.Heads(0x18C15, 0x18C40):
    print(render(ea))

ida_pro.qexit(0)
