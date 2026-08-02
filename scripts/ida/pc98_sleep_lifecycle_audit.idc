#include <idc.idc>

/*
 * Non-destructive IDA Pro 9.4 audit for PC-98 Sleep effect lifetime.
 * Run only on a writable copy of overlay 23; original GAME.EXE/GAME.OVR,
 * extracted overlay and baseline databases remain read-only.
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

static emit_xrefs(out, label, relative_target)
{
  auto base;
  auto target;
  auto x;
  auto count;

  base = get_inf_attr(INF_MIN_EA);
  target = base + relative_target;
  count = 0;
  for (x = get_first_cref_to(target); x != BADADDR;
       x = get_next_cref_to(target, x))
  {
    fprintf(out, "%s_XREF ea=%08X local=%04X type=%d disasm=%s\n",
            label, x, x - base, XrefType(), generate_disasm_line(x, 0));
    count = count + 1;
  }
  fprintf(out, "%s_XREF count=%d\n", label, count);
}

static main()
{
  auto path;
  auto output_path;
  auto out;
  auto base;

  auto_wait();
  path = get_input_file_path();
  if (strstr(path, "overlay-23.bin") == -1)
    qexit(1);
  output_path = "/work/pc98-sleep-lifecycle-overlay23.txt";
  out = fopen(output_path, "w");
  if (out == 0)
    qexit(2);

  base = get_inf_attr(INF_MIN_EA);
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(base, SEGATTR_BITNESS, 0);
  fprintf(out, "input=%s min=%08X max=%08X\n", path, base,
          get_inf_attr(INF_MAX_EA));

  emit_range(out, "EFFECTS_PREFIX", 0x0000, 0x010E);
  emit_range(out, "SPELLOFF", 0x010E, 0x03FE);
  emit_range(out, "CHECKFX", 0x03FE, 0x0E32);
  emit_range(out, "ADDEFFECT_TO_CURE", 0x13D7, 0x1693);
  emit_range(out, "PUTDAMAGE", 0x1FFD, 0x2325);
  emit_range(out, "PUTEFFECT", 0x2325, 0x2419);
  emit_range(out, "STANDUP", 0x251A, 0x2600);
  emit_xrefs(out, "SPELLOFF", 0x010E);
  emit_xrefs(out, "REMOVEFX", 0x158A);
  emit_xrefs(out, "CUREEFFECT", 0x1630);

  fclose(out);
  qexit(0);
}
