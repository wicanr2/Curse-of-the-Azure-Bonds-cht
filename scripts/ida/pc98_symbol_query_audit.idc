#include <idc.idc>

/*
 * Read-only symbol/xref audit for a PC-98 baseline database or executable.
 *
 * The script never renames, patches, or deletes items. It records the original
 * symbol spelling, IDA address space, xrefs and a short continuous instruction
 * window so a later semantic ledger can attach a confidence level without
 * replacing Borland names or linear addresses.
 */

static emit_range(out, label, start, end)
{
  auto ea;
  auto size;
  auto index;

  ea = start;
  while (ea < end)
  {
    size = get_item_size(ea);
    if (size <= 0)
      size = 1;
    fprintf(out, "range=%s ea=%08X bytes=", label, ea);
    for (index = 0; index < size; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(ea + index));
    fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
    ea = ea + size;
  }
}

static emit_symbol(out, symbol_name)
{
  auto ea;
  auto x;
  auto count;

  ea = get_name_ea(BADADDR, symbol_name);
  fprintf(out, "symbol=%s ea=%08X name=%s\n", symbol_name, ea,
          ea == BADADDR ? "<unresolved>" : get_name(ea));
  if (ea == BADADDR)
    return;
  count = 0;
  for (x = get_first_cref_to(ea); x != BADADDR;
       x = get_next_cref_to(ea, x))
  {
    fprintf(out, "xref symbol=%s ea=%08X disasm=%s\n", symbol_name, x,
            generate_disasm_line(x, 0));
    count = count + 1;
  }
  fprintf(out, "xref symbol=%s count=%d\n", symbol_name, count);
}

static emit_data_xrefs(out, label, ea)
{
  auto x;
  auto count;

  fprintf(out, "data_symbol=%s ea=%08X\n", label, ea);
  count = 0;
  for (x = get_first_dref_to(ea); x != BADADDR;
       x = get_next_dref_to(ea, x))
  {
    fprintf(out, "dref symbol=%s ea=%08X type=%d disasm=%s\n", label, x,
            XrefType(), generate_disasm_line(x, 0));
    count = count + 1;
  }
  fprintf(out, "dref symbol=%s count=%d\n", label, count);
}

static emit_segment_layout(out)
{
  auto seg;

  for (seg = get_first_seg(); seg != BADADDR; seg = get_next_seg(seg))
    fprintf(out, "segment start=%08X end=%08X name=%s\n",
            get_segm_start(seg), get_segm_end(seg), get_segm_name(seg));
}

static main()
{
  auto path;
  auto out;
  auto output_path;
  auto start;

  auto_wait();
  path = get_input_file_path();
  output_path = "/work/pc98-symbol-query.txt";
  out = fopen(output_path, "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s min=%08X max=%08X\n", path,
          get_inf_attr(INF_MIN_EA), get_inf_attr(INF_MAX_EA));
  emit_segment_layout(out);

  emit_symbol(out, "COLDFLG");
  emit_symbol(out, "FIGGOODBAD");
  emit_symbol(out, "DOEFFECT");
  emit_symbol(out, "AREAEFFECT");
  emit_symbol(out, "EFFECTNUM");
  emit_symbol(out, "EFFECTPTR");
  emit_symbol(out, "EFFECTPROC");
  emit_symbol(out, "EFFECTREC");
  emit_symbol(out, "LOADEFFECTS");
  emit_symbol(out, "PUTEFFECT");
  emit_symbol(out, "ADDEFFECT");
  emit_symbol(out, "CALLEFFECT");
  emit_symbol(out, "HELPLESSEFFECTS");
  emit_symbol(out, "FIGSPELL");
  emit_symbol(out, "SAVEVS");
  emit_symbol(out, "SAVERESULT");
  emit_symbol(out, "DAMAGE");
  emit_symbol(out, "DAMAGETYPE");
  emit_symbol(out, "SPELLON");

  /* Borland dseg 0C29 is mapped to IDA linear 01C290 in this executable.
   * COLDFLG itself is an absolute Borland enum value (0000:0002), not a
   * dseg global. These drefs therefore audit only resident data symbols; they
   * must not be read as consumers of the COLDFLG constant. */
  emit_data_xrefs(out, "COLDFLG_0C29_0002", 0x01C292);
  emit_data_xrefs(out, "EFFECT_0C29_A02D", 0x0262BD);
  emit_data_xrefs(out, "AREAEFFECT_0C29_A037", 0x0262C7);
  emit_data_xrefs(out, "DOEFFECT_0C29_A044", 0x0262D4);
  start = get_name_ea(BADADDR, "COLDFLG");
  if (start != BADADDR)
    emit_range(out, "COLDFLG_WINDOW", start, start + 0x80);
  start = get_name_ea(BADADDR, "FIGGOODBAD");
  if (start != BADADDR)
    emit_range(out, "FIGGOODBAD_WINDOW", start, start + 0x80);
  start = get_name_ea(BADADDR, "DOEFFECT");
  if (start != BADADDR)
    emit_range(out, "DOEFFECT_WINDOW", start, start + 0x80);
  start = get_name_ea(BADADDR, "AREAEFFECT");
  if (start != BADADDR)
    emit_range(out, "AREAEFFECT_WINDOW", start, start + 0x80);
  fclose(out);
  qexit(0);
}
