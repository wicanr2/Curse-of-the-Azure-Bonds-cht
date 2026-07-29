#include <idc.idc>

/*
 * Native IDC fallback for the PC-9801 SOUND.ROM tone consumer.
 *
 * Load the user's 16 KiB ROM as a raw binary at linear CC000h, then run this
 * script with IDA Pro 9.4. It deliberately prints only addresses and short
 * control tables; no commercial ROM or database belongs in Git.
 */
static main()
{
  auto index;

  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(0xCC000, SEGATTR_BITNESS, 0);
  del_items(0xCC000, DELIT_SIMPLE, 0x4000);
  create_insn(0xCEE08);
  add_func(0xCEE08, BADADDR);
  auto_wait();

  msg("ENTRY CEE08 SETPARABLOCK CF309 SETVOLUME CF41F\n");
  msg("FIELD_COMPLEMENT_SENTINEL CFBEF");
  for (index = 0; index <= 50; index = index + 1) {
    msg(" %02X", get_wide_byte(0xCFBEF + index));
  }
  msg("\nOPERATOR_REGISTER_OFFSETS CFD6E");
  for (index = 0; index < 4; index = index + 1) {
    msg(" %02X", get_wide_byte(0xCFD6E + index));
  }
  msg("\n");
  qexit(0);
}
