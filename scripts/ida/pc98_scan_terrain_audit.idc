#include <idc.idc>

/*
 * Non-destructive IDA Pro 9.4 audit for the PC-98 SCAN terrain globals.
 * Run only against a disposable copy of PC98-GAME.EXE.i64.  The report keeps
 * Borland's original names and linear addresses; this script never renames or
 * comments the database.
 */

static emit_bytes(out, label, ea, count)
{
  auto index;
  fprintf(out, "label=%s ea=%08X segment=%s bytes=", label, ea, get_segm_name(ea));
  for (index = 0; index < count; index = index + 1)
    fprintf(out, "%02X", get_wide_byte(ea + index));
  fprintf(out, "\n");
}

static emit_xrefs(out, label, ea)
{
  auto xref;
  xref = get_first_dref_to(ea);
  while (xref != BADADDR)
  {
    fprintf(out, "xref=%s target=%08X from=%08X type=%d disasm=%s\n",
            label, ea, xref, XrefType(), generate_disasm_line(xref, 0));
    xref = get_next_dref_to(ea, xref);
  }
}

static main()
{
  auto out, data_base, tdef, combatmap, lastsight, index;
  auto_wait();
  out = fopen("/work/pc98-scan-terrain-globals.txt", "w");
  if (out == 0)
    qexit(2);

  /*
   * The pristine database intentionally has no semantic renames.  Borland's
   * appended symbol table independently resolves all three names to segment
   * 0C29.  IDA loads that data segment at linear 1C290h, so preserve both
   * address spaces explicitly instead of writing names into the database.
   */
  data_base = 0x1C290;
  tdef = data_base + 0x48B0;
  combatmap = data_base + 0x9F2C;
  lastsight = data_base + 0x9F30;
  fprintf(out, "input=%s min=%08X max=%08X\n", get_input_file_path(),
          get_inf_attr(INF_MIN_EA), get_inf_attr(INF_MAX_EA));
  fprintf(out, "resolution=borland_segment_0C29 ida_linear_base=%08X segment=%s\n",
          data_base, get_segm_name(data_base));
  fprintf(out, "symbol=TDEF original=0C29:48B0 ea=%08X ida_name=%s\n", tdef, get_name(tdef));
  fprintf(out, "symbol=COMBATMAP original=0C29:9F2C ea=%08X ida_name=%s\n", combatmap, get_name(combatmap));
  fprintf(out, "symbol=LASTSIGHT original=0C29:9F30 ea=%08X ida_name=%s\n", lastsight, get_name(lastsight));

  for (index = 0; index < 65; index = index + 1)
    emit_bytes(out, "TDEF_RECORD", tdef + index * 4, 4);
  emit_bytes(out, "COMBATMAP_POINTER", combatmap, 4);
  emit_bytes(out, "LASTSIGHT_COUNT", lastsight, 1);
  emit_xrefs(out, "TDEF", tdef);
  emit_xrefs(out, "COMBATMAP", combatmap);
  emit_xrefs(out, "LASTSIGHT", lastsight);
  fclose(out);
  qexit(0);
}
