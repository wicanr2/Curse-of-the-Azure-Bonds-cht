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
  /*
   * The ISR dispatches through three register-held near offsets, so normal
   * auto-analysis does not discover these paths.
   */
  del_items(0xCF47A, DELIT_SIMPLE, 0x6CA);
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
  create_insn(0xCF817);
  add_func(0xCF817, BADADDR);
  create_insn(0xCF847);
  add_func(0xCF847, BADADDR);
  create_insn(0xCFA10);
  add_func(0xCFA10, BADADDR);
  create_insn(0xCFA24);
  add_func(0xCFA24, BADADDR);
  create_insn(0xCFA49);
  add_func(0xCFA49, BADADDR);
  create_insn(0xCFA55);
  add_func(0xCFA55, BADADDR);
  create_insn(0xCFA67);
  add_func(0xCFA67, BADADDR);
  create_insn(0xCFA8A);
  add_func(0xCFA8A, BADADDR);
  create_insn(0xCFAB9);
  add_func(0xCFAB9, BADADDR);
  auto_wait();

  msg("ENTRY CEE08 SETPARABLOCK CF309 SETVOLUME CF41F\n");
  msg("TIMER_ISR CF47A COMMON CF4C3 NOTE CF501 LFO CF5F3\n");
  msg("LFO_OUTPUT PITCH CF817 LEVEL CF847 PHASE CFAB9\n");
  msg("LFO_WAVEFORMS 0=CFA49 1=CFA67 2=CFA10 3=CFA8A 4=CFA55 5=CFA24\n");
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
