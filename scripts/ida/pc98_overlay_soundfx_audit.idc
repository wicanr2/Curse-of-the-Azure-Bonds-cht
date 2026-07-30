#include <idc.idc>

/*
 * IDA Pro 9.4 native audit for one raw GAME.OVR code segment.
 *
 * The caller extracts code with cmd/pc98-ovr-audit first, so relocation bytes
 * are excluded. This script recognizes only the exact Turbo Pascal sequence:
 *   PUSH word ptr DS:[selector address]
 *   CALL FAR 0893:0000
 */

static selector_for_address(address)
{
  if (address < 0x4838 || address > 0x4858 ||
      ((address - 0x4838) & 1) != 0) {
    return -2;
  }
  if (address == 0x4838) {
    return 255;
  }
  return ((address - 0x4838) / 2) - 1;
}

static main()
{
  auto ea, end, address, selector, count;
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(get_first_seg(), SEGATTR_BITNESS, 0);
  auto_wait();
  ea = get_segm_start(get_first_seg());
  end = get_segm_end(ea);
  count = 0;
  msg("OVERLAY_SOUNDFX_RANGE start=%08X end=%08X\n", ea, end);
  while (ea + 9 <= end) {
    if (get_wide_byte(ea) == 0xFF &&
        get_wide_byte(ea + 1) == 0x36 &&
        get_wide_byte(ea + 4) == 0x9A &&
        get_wide_word(ea + 5) == 0x0000 &&
        get_wide_word(ea + 7) == 0x0893) {
      address = get_wide_word(ea + 2);
      selector = selector_for_address(address);
      del_items(ea, 0, 9);
      create_insn(ea);
      create_insn(ea + 4);
      msg("OVERLAY_SOUNDFX_CALL push=%08X call=%08X ds=%04X selector=%d push_asm=%s call_asm=%s\n",
          ea, ea + 4, address, selector,
          generate_disasm_line(ea, 0),
          generate_disasm_line(ea + 4, 0));
      count = count + 1;
      ea = ea + 9;
    } else {
      ea = ea + 1;
    }
  }
  msg("OVERLAY_SOUNDFX_CALLS count=%d\n", count);
  qexit(0);
}
