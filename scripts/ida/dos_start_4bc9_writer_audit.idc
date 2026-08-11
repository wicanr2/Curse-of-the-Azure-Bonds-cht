#include <idc.idc>

/*
 * Read-only DOS START.EXE audit for the ECL5 4BC9 work byte.
 *
 * The target is intentionally reported as a DS-relative address only.  The
 * script preserves the original IDA name, bytes, direct data xrefs and
 * caller/function context; it does not rename, patch, or infer a field name.
 */

static emit_context(out, ea)
{
  auto cursor;
  auto start;
  auto end;
  auto index;

  start = ea - 8;
  if (start < get_segm_start(ea))
    start = get_segm_start(ea);
  end = ea + 16;
  if (end > get_segm_end(ea))
    end = get_segm_end(ea);
  fprintf(out, "context from=0x%08X to=0x%08X function=%s\n",
          start, end, get_func_name(ea));
  for (cursor = start; cursor < end; cursor = cursor + 1)
  {
    fprintf(out, "  ea=0x%08X bytes=", cursor);
    for (index = 0; index < 8 && cursor + index < end; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(cursor + index));
    fprintf(out, " asm=%s\n", generate_disasm_line(cursor, 0));
  }
}

static main()
{
  auto input;
  auto seg;
  auto dseg;
  auto start;
  auto end;
  auto target;
  auto xref;
  auto out;
  auto count;

  auto_wait();
  input = get_input_file_path();
  if (strstr(input, "START.EXE") == -1)
    qexit(2);
  dseg = BADADDR;
  for (seg = get_first_seg(); seg != BADADDR; seg = get_next_seg(seg))
  {
    if (get_segm_name(seg) == "dseg")
    {
      dseg = seg;
      break;
    }
  }
  if (dseg == BADADDR)
    qexit(2);
  start = get_segm_start(dseg);
  end = get_segm_end(dseg);
  target = start + 0x4BC9;
  out = fopen("/tmp/dos-start-4bc9-writer.txt", "w");
  if (out == 0)
    qexit(2);

  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=IDA dseg linear EA plus DS-relative offset; direct xrefs only\n");
  fprintf(out, "semantic_status=unknown 4BC9 field name and producer\n");
  fprintf(out, "dseg_start=0x%08X dseg_end=0x%08X target_ds_offset=0x4BC9 target_ea=0x%08X\n",
          start, end, target);
  if (target < start || target >= end)
  {
    fprintf(out, "target_in_dseg=0\n");
    fclose(out);
    qexit(0);
  }
  fprintf(out, "target_in_dseg=1 name=%s item_size=%d bytes=",
          get_name(target), get_item_size(target));
  for (seg = 0; seg < 16; seg = seg + 1)
    fprintf(out, "%02X", get_wide_byte(target + seg));
  fprintf(out, " disasm=%s\n", generate_disasm_line(target, 0));
  count = 0;
  for (xref = get_first_dref_to(target); xref != BADADDR;
       xref = get_next_dref_to(target, xref))
  {
    fprintf(out, "dref from=0x%08X type=%d function=%s disasm=%s\n",
            xref, XrefType(), get_func_name(xref), generate_disasm_line(xref, 0));
    emit_context(out, xref);
    count = count + 1;
  }
  fprintf(out, "dref_count=%d\n", count);
  fclose(out);
  qexit(0);
}
