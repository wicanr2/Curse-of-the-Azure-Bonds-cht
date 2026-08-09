#include <idc.idc>

/*
 * Read-only DOS START.EXE dseg offset audit.
 *
 * GAME.OVR overlay code uses DS-relative operands that are not represented by
 * the extracted overlay database.  This script maps only the requested DS
 * offsets into the resident IDA dseg, then reports original names, bytes and
 * direct data xrefs.  It does not rename, patch, or import semantic guesses.
 */

static emit_target(out, dseg_start, dseg_end, offset)
{
  auto ea;
  auto x;
  auto count;
  auto size;
  auto index;

  ea = dseg_start + offset;
  fprintf(out, "target ds_offset=0x%04X ea=0x%08X in_dseg=%d name=%s\n",
          offset, ea, ea >= dseg_start && ea < dseg_end ? 1 : 0,
          ea >= dseg_start && ea < dseg_end ? get_name(ea) : "<out-of-range>");
  if (ea < dseg_start || ea >= dseg_end)
    return;

  size = get_item_size(ea);
  if (size <= 0)
    size = 1;
  fprintf(out, "bytes ds_offset=0x%04X size=%d data=", offset, size);
  for (index = 0; index < size && index < 16; index = index + 1)
    fprintf(out, "%02X", get_wide_byte(ea + index));
  fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));

  count = 0;
  for (x = get_first_dref_to(ea); x != BADADDR;
       x = get_next_dref_to(ea, x))
  {
    fprintf(out, "dref ds_offset=0x%04X from=0x%08X type=%d disasm=%s\n",
            offset, x, XrefType(), generate_disasm_line(x, 0));
    count = count + 1;
  }
  fprintf(out, "dref ds_offset=0x%04X count=%d\n", offset, count);
}

static main()
{
  auto input;
  auto seg;
  auto dseg;
  auto start;
  auto end;
  auto out;

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
  out = fopen("/tmp/dos-start-dseg-offsets.txt", "w");
  if (out == 0)
    qexit(2);

  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=IDA dseg linear EA plus DS-relative offset; direct xrefs only\n");
  fprintf(out, "semantic_status=unknown DS work-cell names; bytes/xrefs only\n");
  fprintf(out, "dseg_start=0x%08X dseg_end=0x%08X\n", start, end);
  emit_target(out, start, end, 0x7206);
  emit_target(out, start, end, 0x720F);
  emit_target(out, start, end, 0x7210);
  emit_target(out, start, end, 0x7211);
  emit_target(out, start, end, 0x7212);
  emit_target(out, start, end, 0x7213);
  emit_target(out, start, end, 0x8B5E);
  emit_target(out, start, end, 0x8B62);
  emit_target(out, start, end, 0x8B65);
  emit_target(out, start, end, 0x8B67);
  emit_target(out, start, end, 0x8B68);
  emit_target(out, start, end, 0x8B6A);
  fclose(out);
  qexit(0);
}
