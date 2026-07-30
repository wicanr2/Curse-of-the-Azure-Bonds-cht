#include <idc.idc>

/*
 * IDA Pro 9.4 native audit for GAME.EXE music transitions and PC-speaker SFX.
 *
 * Run against the user's exact GAME.EXE database. Commercial bytes,
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
    /*
     * GAME.EXE has segment-dependent MZ mappings. Native IDC does not expose
     * ida_loader.get_fileregion_offset(), so do not print a guessed offset;
     * the versioned raw-byte auditor supplies exact file anchors separately.
     */
    msg("%08X bytes=", ea);
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
    msg("CALLER ea=%08X asm=%s\n",
        ea, generate_disasm_line(ea, 0));
    ea = get_next_cref_to(target, ea);
  }
}

static dump_bytes(label, start, size)
{
  auto index;
  msg("BYTES label=%s start=%08X size=%d hex=", label, start, size);
  for (index = 0; index < size; index = index + 1) {
    msg("%02X", get_wide_byte(start + index));
  }
  msg("\n");
}

static main()
{
  auto_wait();
  dump_range("GAME_SOUNDFX", 0x18930, 0x18A3D);
  dump_bytes("GAME_SOUNDFX_SEQUENCE_TABLE", 0x1DCCC, 16 * 20 * 2);
  dump_range("MSCPLAY", 0x18A44, 0x18A8E);
  dump_range("MSCSTOP", 0x18A8E, 0x18AA7);
  dump_range("BGMPLAY", 0x18AA7, 0x18B8F);
  dump_range("BORLAND_DELAY_SOUND_NOSOUND", 0x19259, 0x1929E);
  dump_range("PC98_PORT37_PULSE", 0x19D1E, 0x19D6A);
  dump_callers("MSCPLAY", 0x18A44);
  dump_callers("MSCSTOP", 0x18A8E);
  dump_callers("BORLAND_DELAY", 0x19259);
  dump_callers("BORLAND_SOUND", 0x19286);
  dump_callers("BORLAND_NOSOUND", 0x19292);
  qexit(0);
}
