#include <idc.idc>

static dump_word_symbol(name, ea)
{
  msg("SOUNDFX_SYMBOL name=%s ea=%08X value=%u\n",
      name, ea, get_wide_word(ea));
}

static main()
{
  auto_wait();
  // Borland symbols use DS 0C29h. This MZ database has IDA load base 10000h,
  // so DS:4838h maps to 20AC8h; entries are consecutive WORDs.
  dump_word_symbol("SOUNDHALT",    0x20AC8);
  dump_word_symbol("SOUNDOFF",     0x20ACA);
  dump_word_symbol("SOUNDON",      0x20ACC);
  dump_word_symbol("CASTFX",       0x20ACE);
  dump_word_symbol("MISSFX",       0x20AD0);
  dump_word_symbol("SPELLHITFX",   0x20AD2);
  dump_word_symbol("DEADFX",       0x20AD4);
  dump_word_symbol("WHISTLEFX",    0x20AD6);
  dump_word_symbol("HITFX",        0x20AD8);
  dump_word_symbol("LIGHTNINGFX",  0x20ADA);
  dump_word_symbol("SWISHFX",      0x20ADC);
  dump_word_symbol("PADFX",        0x20ADE);
  dump_word_symbol("FIREBALLFX",   0x20AE0);
  dump_word_symbol("ARROWFX",      0x20AE2);
  dump_word_symbol("OVERTUREFX",   0x20AE4);
  dump_word_symbol("COMBATFX",     0x20AE6);
  dump_word_symbol("CRASHFX",      0x20AE8);
  qexit(0);
}
