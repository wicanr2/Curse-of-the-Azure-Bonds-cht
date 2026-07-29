"""IDA Pro batch audit for the incomplete PC-98 MSCDRV 7Eh/D2h bridge.

Run only against the user's local MSCDRV.EXE database. The commercial binary,
generated IDA database, and missing-sector reconstruction must not be committed.
"""

import ida_auto
import ida_bytes
import ida_loader
import ida_nalt
import ida_pro
import idautils
import idc


RANGES = {
    "VECTOR_7E_HANDLER": (0x10080, 0x100CC),
    "INSTALL_VECTOR_7E": (0x100E3, 0x100F6),
    "SELECT_TRACK": (0x1021E, 0x10253),
    "STOP_PLAYBACK": (0x1037E, 0x10410),
    "INSTALL_INT_D2": (0x110CA, 0x110EA),
    "INITIALIZE_INT_D2": (0x110EA, 0x11161),
    "INT_D2_CLIENT_WRAPPERS": (0x11172, 0x112B5),
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
for label, (start, end) in RANGES.items():
    print(f"RANGE label={label} start={start:08X} end={end:08X}")
    for ea in idautils.Heads(start, end):
        print(render(ea))

for needle_label, needle in (
    ("VECTOR_7E", 0x7E),
    ("INT_D2", 0xD2),
    ("DOS_GET_VECTOR", 0x35),
    ("DOS_SET_VECTOR", 0x25),
):
    print(f"IMMEDIATE_SCAN label={needle_label} value={needle:02X}")
    for segment_ea in idautils.Segments():
        segment_end = idc.get_segm_end(segment_ea)
        for ea in idautils.Heads(segment_ea, segment_end):
            if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
                continue
            if any(idc.get_operand_value(ea, index) == needle for index in range(2)):
                print(render(ea))

ida_pro.qexit(0)
