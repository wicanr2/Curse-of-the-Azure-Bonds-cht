"""IDA Pro batch audit for MSCDRV's PC-9801 Sound BIOS clients.

Run only against the user's local MSCDRV.EXE database.  This script records
the title-specific callers and raw instruction bytes; NEC's official command
names and register contracts remain documented separately in spec 365.
"""

import ida_auto
import ida_bytes
import ida_funcs
import ida_loader
import ida_nalt
import ida_pro
import idautils
import idc


FUNCTIONS = {
    "DIRECT_YM2203_WRITE": 0x11075,
    "DIRECT_YM2203_READ": 0x110AB,
    "SBIOS_INITIALIZE": 0x1115D,
    "SBIOS_CLEAR": 0x11172,
    "SBIOS_READREG": 0x11184,
    "SBIOS_WRITEREG": 0x111A1,
    "SBIOS_SETTOUCH": 0x111B4,
    "SBIOS_NOTE": 0x111CA,
    "SBIOS_SETLENGTH": 0x111DD,
    "SBIOS_SETREGBUFFER": 0x111E9,
    "SBIOS_SETPARABLOCK": 0x1120D,
    "SBIOS_READPARA": 0x11220,
    "SBIOS_WRITEPARA": 0x11238,
    "SBIOS_ALLSTOP": 0x1123D,
    "SBIOS_CONTPLAY": 0x11242,
    "SBIOS_HOLDSTATE": 0x11252,
    "SBIOS_MODUON": 0x11262,
    "SBIOS_MODUOFF": 0x11272,
    "SBIOS_SETINTCOND": 0x11288,
    "SBIOS_SETVOLUME": 0x1129D,
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

for label, start in FUNCTIONS.items():
    function = ida_funcs.get_func(start)
    if function is None:
        print(f"FUNCTION label={label} start={start:08X} missing=1")
        continue
    print(
        f"FUNCTION label={label} start={function.start_ea:08X} "
        f"end={function.end_ea:08X}"
    )
    for ea in idautils.Heads(function.start_ea, function.end_ea):
        print(render(ea))
    for caller in idautils.CodeRefsTo(function.start_ea, False):
        owner = ida_funcs.get_func(caller)
        owner_start = owner.start_ea if owner else idc.BADADDR
        print(
            f"CALLER target={label} ea={caller:08X} "
            f"function={owner_start:08X} {render(caller)}"
        )

print("INTERRUPT_SCAN vector=D2")
for segment_ea in idautils.Segments():
    for ea in idautils.Heads(segment_ea, idc.get_segm_end(segment_ea)):
        if (
            ida_bytes.is_code(ida_bytes.get_full_flags(ea))
            and idc.print_insn_mnem(ea).lower() == "int"
            and idc.get_operand_value(ea, 0) == 0xD2
        ):
            function = ida_funcs.get_func(ea)
            owner_start = function.start_ea if function else idc.BADADDR
            print(f"INT_D2 function={owner_start:08X} {render(ea)}")

ida_pro.qexit(0)
