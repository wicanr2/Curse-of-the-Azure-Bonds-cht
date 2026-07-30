#include <idc.idc>

/*
 * IDA Pro 9.4 native audit for the PC-98 MSCDRV timer setup and dispatch.
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
    /* MZ image load base 10000h and 200h-byte file header. */
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

static main()
{
  auto_wait();
  dump_range("RESIDENT_HELPERS", 0x10020, 0x10080);
  dump_range("TRACK_TIMER_DISPATCH", 0x10175, 0x1021E);
  dump_range("TIMER_INITIALIZE", 0x10CD3, 0x10D3A);
  dump_range("TIMER_STATE_UPDATE", 0x10D3A, 0x10DAA);
  dump_range("TIMER_STOP", 0x10DAA, 0x10DF0);
  dump_range("DRIVER_TIMER_HOOK", 0x10EE0, 0x1102A);
  dump_range("DIRECT_OPN_WRITE", 0x1102A, 0x110C4);
  dump_range("TEMPO_AND_INTERRUPT_ADAPTERS", 0x111E2, 0x11299);
  dump_callers("TRACK_TIMER_DISPATCH", 0x10175);
  dump_callers("TEMPO_REGISTER_ADAPTER", 0x111E2);
  dump_callers("SETINTCOND_ADAPTER", 0x11277);
  dump_callers("SETINTCOND", 0x11288);
  qexit(0);
}
