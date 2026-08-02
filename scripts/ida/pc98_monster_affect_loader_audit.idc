#include <idc.idc>

/*
 * Non-destructive IDA Pro 9.4 audit for PC-98 monster affect records.
 *
 * Run only on copies of code-only overlays emitted by pc98-ovr-audit:
 *   overlay 12 EFFPROCS, overlay 16 LOADSAVE, overlay 23 EFFECTS.
 * Each invocation writes a separate report under /work. The original GAME.EXE,
 * GAME.OVR and baseline databases remain read-only.
 */

static emit_range(out, label, relative_start, relative_end)
{
  auto base;
  auto start;
  auto end;
  auto ea;
  auto size;
  auto index;

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

static decode_all()
{
  auto base;
  auto end;
  auto ea;
  auto size;

  base = get_inf_attr(INF_MIN_EA);
  end = get_inf_attr(INF_MAX_EA);
  del_items(base, DELIT_SIMPLE, end - base);
  ea = base;
  while (ea < end)
  {
    size = create_insn(ea);
    if (size <= 0)
      size = 1;
    ea = ea + size;
  }
}

static main()
{
  auto path;
  auto out;
  auto output_path;
  auto base;
  auto target;
  auto x;
  auto count;

  auto_wait();
  path = get_input_file_path();
  base = get_inf_attr(INF_MIN_EA);
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(base, SEGATTR_BITNESS, 0);

  if (strstr(path, "overlay-12.bin") != -1)
    output_path = "/work/pc98-monster-affect-overlay12.txt";
  else if (strstr(path, "overlay-16.bin") != -1)
    output_path = "/work/pc98-monster-affect-overlay16.txt";
  else if (strstr(path, "overlay-22.bin") != -1)
    output_path = "/work/pc98-monster-affect-overlay22.txt";
  else if (strstr(path, "overlay-23.bin") != -1)
    output_path = "/work/pc98-monster-affect-overlay23.txt";
  else
    qexit(1);

  out = fopen(output_path, "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s min=%08X max=%08X\n", path, base,
          get_inf_attr(INF_MAX_EA));

  if (strstr(path, "overlay-12.bin") != -1)
  {
    emit_range(out, "PROTECTED_HANDLER", 0x001B, 0x003C);
    emit_range(out, "EFFECT_4F_RESOLVED_ENTRY", 0x19B3, 0x1A20);
    emit_range(out, "MAGIC_RESISTANCE_HANDLERS", 0x2396, 0x2420);
    emit_range(out, "EFFECT_70_RESOLVED_ENTRY", 0x249D, 0x24D0);
    emit_range(out, "EFFECT_87_RESOLVED_ENTRY", 0x2B87, 0x2BC0);
    emit_range(out, "EFFECT_18_RESOLVED_ENTRY", 0x2ECB, 0x2ED4);
    emit_range(out, "INITEFFPROX", 0x2ED4, 0x3604);
  }
  else if (strstr(path, "overlay-16.bin") != -1)
  {
    emit_range(out, "LOADMONSTER_EFFECT_COPY", 0x3BFF, 0x3C8D);
  }
  else if (strstr(path, "overlay-22.bin") != -1)
  {
    emit_range(out, "THROWS_LIGHTNING_HANDLER", 0x62D7, 0x6374);
    emit_range(out, "INITSPELLS_EFFECT_TABLE_TAIL", 0x6C60, 0x6C8C);
  }
  else
  {
    decode_all();
    emit_range(out, "CALLEFFECT", 0x00C9, 0x010E);
    emit_range(out, "EFFECT_PHASE_CALLER_0184", 0x0110, 0x01B0);
    emit_range(out, "CHECKFX_TYPE_2_3", 0x03FE, 0x0600);
    emit_range(out, "EFFECT_PHASE_CALLERS_0F9C_104E", 0x0F20, 0x1070);
    emit_range(out, "PUTEFFECT", 0x2325, 0x2419);
    emit_range(out, "EFFECT_PHASE_CALLER_24DA", 0x2440, 0x2520);
    target = base + 0x00C9;
    count = 0;
    for (x = get_first_cref_to(target); x != BADADDR;
         x = get_next_cref_to(target, x))
    {
      fprintf(out, "CALLEFFECT_XREF ea=%08X local=%04X type=%d disasm=%s\n",
              x, x - base, XrefType(), generate_disasm_line(x, 0));
      count = count + 1;
    }
    fprintf(out, "CALLEFFECT_XREF count=%d\n", count);
    emit_range(out, "ADDEFFECT", 0x13D7, 0x1486);
  }
  fclose(out);
  qexit(0);
}
