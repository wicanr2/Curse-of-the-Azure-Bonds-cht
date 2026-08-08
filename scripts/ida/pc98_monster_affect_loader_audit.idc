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
    emit_range(out, "EFFECT_35_RESOLVED_ENTRY", 0x0060, 0x00C0);
    emit_range(out, "EFFECT_01_TO_0A_HANDLERS", 0x009E, 0x0300);
    /*
     * Keep the effect work-cell consumers adjacent to the handler evidence.
     * These are intentionally bounded windows around raw A02C/A039 references;
     * the flat listing is a search aid only, while the .i64 xref graph remains
     * the primary relation evidence.
     */
    emit_range(out, "EFFECT_WORK_CONSUMER_0400_0500", 0x0400, 0x0500);
    emit_range(out, "EFFECT_WORK_CONSUMER_0700_0770", 0x0700, 0x0770);
    emit_range(out, "EFFECT_WORK_CONSUMER_0980_09D0", 0x0980, 0x09D0);
    emit_range(out, "EFFECT_WORK_CONSUMER_0B90_0C20", 0x0B90, 0x0C20);
    emit_range(out, "EFFECT_WORK_CONSUMER_0F90_1010", 0x0F90, 0x1010);
    emit_range(out, "EFFECT_WORK_CONSUMER_11D0_1230", 0x11D0, 0x1230);
    emit_range(out, "EFFECT_WORK_CONSUMER_1250_12C0", 0x1250, 0x12C0);
    emit_range(out, "EFFECT_WORK_CONSUMER_13C0_1410", 0x13C0, 0x1410);
    emit_range(out, "EFFECT_WORK_CONSUMER_16D0_1730", 0x16D0, 0x1730);
    emit_range(out, "EFFECT_WORK_CONSUMER_17F0_1850", 0x17F0, 0x1850);
    emit_range(out, "EFFECT_WORK_CONSUMER_1CF0_1D50", 0x1CF0, 0x1D50);
    emit_range(out, "EFFECT_WORK_CONSUMER_2080_2180", 0x2080, 0x2180);
    emit_range(out, "EFFECT_WORK_CONSUMER_28D0_2910", 0x28D0, 0x2910);
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
    emit_range(out, "EFFECT_CHECK_HELPER_0268_03FE", 0x0268, 0x03FE);
    emit_range(out, "CHECKFX_TYPE_2_3", 0x03FE, 0x0600);
    emit_range(out, "CHECKFX_TAIL_0600_0CE7", 0x0600, 0x0CE7);
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
