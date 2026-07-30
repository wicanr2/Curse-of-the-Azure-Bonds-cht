#include <idc.idc>

/*
 * IDA Pro 9.4 native audit for Burial Glen Daemir ECL work-cell consumers.
 *
 * Run against DOS START.EXE and a separately loaded 16-bit GAME.OVR.
 * Literal operands alone do not prove semantics; this script records the
 * instruction context for 4CBAh, 4CBBh, and 4CC0h so callers/consumers can be
 * cross-checked against ECL bytes and runtime behavior.
 */

static is_target(value)
{
  value = value & 0xFFFF;
  return value == 0x4CBA || value == 0x4CBB || value == 0x4CC0;
}

static main()
{
  auto ea;
  auto end;
  auto op;
  auto value;
  auto hits;
  auto raw_hits;

  auto_wait();
  ea = get_inf_attr(INF_MIN_EA);
  end = get_inf_attr(INF_MAX_EA);
  hits = 0;
  raw_hits = 0;
  msg("DOS_DAEMIR_WORK_CELLS input=%s min=%08X max=%08X\n",
      get_input_file_path(), ea, end);

  while (ea != BADADDR && ea < end)
  {
    if (is_code(get_full_flags(ea)))
    {
      for (op = 0; op < 6; op = op + 1)
      {
        value = get_operand_value(ea, op);
        if (is_target(value))
        {
          msg("DOS_DAEMIR_REF ea=%08X target=%04X op=%d text=%s disasm=%s\n",
              ea, value & 0xFFFF, op, print_operand(ea, op),
              generate_disasm_line(ea, 0));
          hits = hits + 1;
        }
      }
    }
    ea = next_head(ea, end);
  }

  for (ea = get_inf_attr(INF_MIN_EA); ea + 1 < end; ea = ea + 1)
  {
    value = get_wide_word(ea);
    if (is_target(value))
    {
      msg("DOS_DAEMIR_RAW_LE16 ea=%08X target=%04X bytes=%02X %02X\n",
          ea, value & 0xFFFF, get_wide_byte(ea), get_wide_byte(ea + 1));
      raw_hits = raw_hits + 1;
    }
  }

  msg("DOS_DAEMIR_WORK_CELLS instruction_hits=%d raw_hits=%d\n",
      hits, raw_hits);
  qexit(0);
}
