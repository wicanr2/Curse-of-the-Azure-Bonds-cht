#include <idc.idc>

/*
 * Non-destructive IDA Pro 9.4 audit for the PC-98 monster lightning phase.
 * Run only on code-only overlay copies emitted by pc98-ovr-audit:
 *   overlay 9  COMPTACT turn-phase CHECKFX type 14 caller
 *   overlay 22 SPELLS effect 84 handler and Lightning Bolt line helpers
 * Reports are written to /work; source overlays and baseline databases stay
 * read-only and no semantic names replace original addresses.
 */

static emit_range(out, label, relative_start, relative_end)
{
  auto base, start, end, ea, size, index;
  base = get_inf_attr(INF_MIN_EA);
  start = base + relative_start;
  end = base + relative_end;
  if (end > get_inf_attr(INF_MAX_EA))
    end = get_inf_attr(INF_MAX_EA);
  del_items(start, DELIT_SIMPLE, end - start);
  ea = start;
  while (ea < end)
  {
    size = create_insn(ea);
    if (size <= 0)
      size = 1;
    fprintf(out, "label=%s ea=%08X local=%04X bytes=", label, ea, ea - base);
    for (index = 0; index < size; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(ea + index));
    fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
    ea = ea + size;
  }
}

static main()
{
  auto path, output_path, out, base, end, ea, size;
  auto_wait();
  path = get_input_file_path();
  base = get_inf_attr(INF_MIN_EA);
  end = get_inf_attr(INF_MAX_EA);
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(base, SEGATTR_BITNESS, 0);
  del_items(base, DELIT_SIMPLE, end - base);
  ea = base;
  while (ea < end)
  {
    size = create_insn(ea);
    if (size <= 0)
      size = 1;
    ea = ea + size;
  }

  if (strstr(path, "overlay-09.bin") != -1)
    output_path = "/work/pc98-monster-lightning-overlay09.txt";
  else if (strstr(path, "overlay-22.bin") != -1)
    output_path = "/work/pc98-monster-lightning-overlay22.txt";
  else
    qexit(1);

  out = fopen(output_path, "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s min=%08X max=%08X\n", path, base, end);

  if (strstr(path, "overlay-09.bin") != -1)
  {
    emit_range(out, "TURN_SPECIAL_EFFECT_PHASE", 0x0D80, 0x0E40);
  }
  else
  {
    emit_range(out, "LIGHTNING_LINE_DAMAGE", 0x5F70, 0x61B0);
    emit_range(out, "EFFECT_84_THROWS_LIGHTNING", 0x62D7, 0x63C0);
    emit_range(out, "EFFECT_84_TABLE_WRITER", 0x6C60, 0x6C8C);
  }
  fclose(out);
  qexit(0);
}
