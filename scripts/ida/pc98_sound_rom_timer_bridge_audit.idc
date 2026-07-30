#include <idc.idc>

/*
 * IDA Pro 9.4 native audit for the PC-98 SOUND.ROM YM2203 timer bridge.
 *
 * Load the user's exact 16 KiB ROM as 8086/16-bit raw binary at CC000h.
 * Commercial bytes, databases, and generated logs stay local.
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
    msg("%08X file=%08X bytes=", ea, ea - 0xCC000);
    for (index = 0; index < size; index = index + 1) {
      msg("%02X", get_wide_byte(ea + index));
    }
    msg(" asm=%s\n", line);
    ea = next_head(ea, end);
  }
}

static main()
{
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(0xCC000, SEGATTR_BITNESS, 0);

  /*
   * The ISR dispatches through register-held near offsets. Recreate the
   * entries that ordinary auto-analysis can leave as data.
   */
  create_insn(0xCF47A);
  add_func(0xCF47A, 0xCF4C3);
  create_insn(0xCF4C3);
  add_func(0xCF4C3, BADADDR);
  create_insn(0xCF501);
  add_func(0xCF501, BADADDR);
  create_insn(0xCF5F3);
  add_func(0xCF5F3, BADADDR);
  create_insn(0xCF7A8);
  add_func(0xCF7A8, BADADDR);
  auto_wait();

  dump_range("TIMER_ISR", 0xCF47A, 0xCF4C3);
  dump_range("TIMER_COMMON_RETURN", 0xCF4C3, 0xCF501);
  dump_range("TIMER_A_NOTE_LENGTH", 0xCF501, 0xCF5F3);
  dump_range("TIMER_B_LFO", 0xCF5F3, 0xCF7A8);
  dump_range("TIMER_B_OUTPUT", 0xCF7A8, 0xCF817);
  qexit(0);
}
