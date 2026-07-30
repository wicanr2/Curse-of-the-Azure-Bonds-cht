#include <idc.idc>

/*
 * IDA Pro 9.4 native audit for MSCDRV transition, fade, and internal SFX.
 *
 * Run against the user's exact MSCDRV.EXE database. Commercial bytes,
 * databases, and generated logs stay local.
 */

static dump_range(label, start, end)
{
  auto ea, size, index, line;
  msg("RANGE label=%s start=%08X end=%08X\n", label, start, end);
  ea = start;
  while (ea != BADADDR && ea < end) {
    size = get_item_size(ea);
    if (size <= 0) {
      size = 1;
    }
    line = generate_disasm_line(ea, 0);
    msg("%08X file=%08X bytes=", ea, ea - 0x10000 + 0x200);
    for (index = 0; index < size; index = index + 1) {
      msg("%02X", get_wide_byte(ea + index));
    }
    msg(" asm=%s\n", line);
    ea = next_head(ea, end);
  }
}

static dump_callers(label, target)
{
  auto ea;
  msg("CALLERS label=%s target=%08X\n", label, target);
  ea = get_first_cref_to(target);
  while (ea != BADADDR) {
    msg("CALLER ea=%08X file=%08X asm=%s\n",
        ea, ea - 0x10000 + 0x200, generate_disasm_line(ea, 0));
    ea = get_next_cref_to(target, ea);
  }
}

static dump_data_xrefs(label, target)
{
  auto ea;
  msg("DATA_XREFS label=%s target=%08X\n", label, target);
  ea = get_first_dref_to(target);
  while (ea != BADADDR) {
    msg("DATA_XREF ea=%08X file=%08X asm=%s\n",
        ea, ea - 0x10000 + 0x200, generate_disasm_line(ea, 0));
    ea = get_next_dref_to(target, ea);
  }
}

static main()
{
  auto_wait();
  dump_range("TRACK_DISPATCH", 0x10175, 0x1021E);
  dump_range("QUEUE_TRACK", 0x1021E, 0x10253);
  dump_range("INITIALIZE_TRACK", 0x10253, 0x1037E);
  dump_range("STREAM_FADE_PREFIX", 0x10410, 0x104F8);
  dump_range("FM_STREAM_END", 0x105A0, 0x10620);
  dump_range("PSG_STREAM_END", 0x10920, 0x10990);
  dump_range("TRACK_STOP_AND_AUDIO_RESET", 0x10C90, 0x10D3A);
  dump_range("INTERNAL_SFX_START", 0x10D3A, 0x10DAA);
  dump_range("INTERNAL_SFX_STEP", 0x10DAA, 0x10EE0);
  dump_callers("TRACK_DISPATCH", 0x10175);
  dump_callers("QUEUE_TRACK", 0x1021E);
  dump_callers("INITIALIZE_TRACK", 0x10253);
  dump_callers("STREAM_INTERPRETER", 0x10410);
  dump_callers("INTERNAL_SFX_START", 0x10D3A);
  dump_callers("INTERNAL_SFX_STEP", 0x10DAA);
  dump_data_xrefs("FADE_COUNTER_DS_0794", 0x13A64);
  dump_data_xrefs("RAW_SFX_REQUEST_DS_0796", 0x13A66);
  dump_data_xrefs("RAW_SFX_ACTIVE_DS_0798", 0x13A68);
  dump_data_xrefs("ACTIVE_TRACK_DS_0A94", 0x13D64);
  dump_data_xrefs("PENDING_TRACK_DS_0A98", 0x13D68);
  qexit(0);
}
