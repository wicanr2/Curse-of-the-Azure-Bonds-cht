#include <idc.idc>

/*
 * IDA Pro 9.4 audit for the PC-98 combat path that may consume the
 * PARTYBASE+01BBh work byte written by Princess Daemir's ECL event.
 *
 * Input identity:
 *   GAME.EXE SHA-256
 *   8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0
 *
 * Borland debug symbols provide these overlay-local anchors:
 *   POSTCOM overlay 5:  DOPOSTCOMBAT 1775h
 *   EFFECTS overlay 23: TRYTOHIT 11C4h, ATTEMPTTOHIT 122Ch
 *   GENERIC overlay 24: STRHITBONUS 1530h
 *   GENTHAC0      0C29:6ECC -> IDA EA 2315Ch (data pointer)
 *   ROLLTOHIT     0C29:A039 -> IDA EA 262C9h (data)
 *   PARTY         0C29:7F05 -> IDA EA 24195h (pointer)
 *
 * A numeric operand hit is only a candidate. The emitted instruction context
 * must still be tied to PARTY and the attack formula before naming 4CBBh.
 */

static is_candidate(value)
{
  value = value & 0xFFFF;
  return value == 0x01BB || value == 0x01BA ||
         value == 0x4CBB || value == 0x4B00 ||
         value == 0x06E0 || value == 0x06E2 ||
         value == 0x0371 || value == 0x6ECC ||
         value == 0x7F05 || value == 0xA039;
}

static dump_range(label, start, end)
{
  auto ea;
  auto size;
  auto index;

  msg("PC98_DAEMIR_RANGE label=%s start=%08X end=%08X\n", label, start, end);
  del_items(start, DELIT_SIMPLE, end - start);
  ea = start;
  while (ea < end)
  {
    if (!is_code(get_full_flags(ea)))
    {
      create_insn(ea);
    }
    size = get_item_size(ea);
    if (size <= 0)
    {
      size = 1;
    }
    msg("PC98_DAEMIR_INSN label=%s ea=%08X bytes=", label, ea);
    for (index = 0; index < size; index = index + 1)
    {
      msg("%02X", get_wide_byte(ea + index));
    }
    msg(" disasm=%s\n", generate_disasm_line(ea, 0));
    ea = ea + size;
  }
}

static main()
{
  auto ea;
  auto end;
  auto op;
  auto value;
  auto hits;
  auto base;

  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(get_first_seg(), SEGATTR_BITNESS, 0);
  auto_wait();
  base = get_segm_start(get_first_seg());
  msg("PC98_DAEMIR_HIT_CONSUMER input=%s min=%08X max=%08X\n",
      get_input_file_path(), get_inf_attr(INF_MIN_EA), get_inf_attr(INF_MAX_EA));

  /*
   * These ranges are overlay-local. Run the script separately against the
   * verified extracted overlay-05.bin, overlay-23.bin, and overlay-24.bin.
   * Ranges that belong to another overlay are printed only as negative
   * controls and must not be assigned that routine's symbol.
   */
  if (get_inf_attr(INF_MAX_EA) >= base + 0x19AA)
  {
    dump_range("LOCAL_0300_03C0", base + 0x0300, base + 0x03C0);
    dump_range("LOCAL_1775_19AA", base + 0x1775, base + 0x19AA);
  }
  dump_range("LOCAL_11C4_122C", base + 0x11C4, base + 0x122C);
  dump_range("LOCAL_122C_12D8", base + 0x122C, base + 0x12D8);
  dump_range("LOCAL_1530_15D3", base + 0x1530, base + 0x15D3);

  ea = get_inf_attr(INF_MIN_EA);
  end = get_inf_attr(INF_MAX_EA);
  hits = 0;
  while (ea != BADADDR && ea < end)
  {
    if (is_code(get_full_flags(ea)))
    {
      for (op = 0; op < 6; op = op + 1)
      {
        value = get_operand_value(ea, op);
        if (is_candidate(value))
        {
          msg("PC98_DAEMIR_CANDIDATE ea=%08X value=%04X op=%d text=%s disasm=%s\n",
              ea, value & 0xFFFF, op, print_operand(ea, op),
              generate_disasm_line(ea, 0));
          hits = hits + 1;
        }
      }
    }
    ea = next_head(ea, end);
  }
  msg("PC98_DAEMIR_HIT_CONSUMER candidate_hits=%d\n", hits);
  qexit(0);
}
