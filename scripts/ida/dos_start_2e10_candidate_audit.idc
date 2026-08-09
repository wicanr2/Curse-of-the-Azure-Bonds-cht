#include <idc.idc>

/*
 * Non-destructive IDA Pro 9.4 audit for the DOS resident START.EXE database.
 *
 * The ECL CALL operand 2E10h is an ECL/external-routine address, not
 * automatically an IDA linear address.  This report deliberately emits raw
 * little-endian candidates in every segment and, separately, IDA xrefs to
 * the two common linear interpretations.  It never renames, patches, or
 * deletes the input database.  A numeric hit remains a candidate until its
 * caller, address space, and runtime consumer are closed.
 */

static emit_xrefs(out, label, target)
{
  auto x;
  auto count;

  count = 0;
  if (target == BADADDR)
  {
    fprintf(out, "xref_target=%s ea=BADADDR count=0\n", label);
    return;
  }
  for (x = get_first_cref_to(target); x != BADADDR;
       x = get_next_cref_to(target, x))
  {
    fprintf(out, "xref_target=%s ea=%08X name=%s disasm=%s\n", label, x,
            get_name(x), generate_disasm_line(x, 0));
    count = count + 1;
  }
  fprintf(out, "xref_target=%s ea=%08X count=%d\n", label, target, count);
}

static emit_raw_window(out, ea, end)
{
  auto index;
  fprintf(out, "raw_window ea=%08X bytes=", ea);
  for (index = ea; index < end; index = index + 1)
    fprintf(out, "%02X", get_wide_byte(index));
  fprintf(out, "\n");
}

static emit_candidate(out, ea, segment_name, target)
{
  auto head;
  auto item_size;
  auto flags;
  auto x;
  auto count;
  auto index;

  head = get_item_head(ea);
  item_size = get_item_size(head);
  if (item_size <= 0)
    item_size = 1;
  flags = get_flags(head);
  fprintf(out, "candidate segment=%s raw_ea=%08X target=0x%04X item_head=%08X item_size=%d code=%d name=%s disasm=%s bytes=",
          segment_name, ea, target, head, item_size, is_code(flags),
          get_name(head), generate_disasm_line(head, 0));
  for (index = 0; index < item_size && index < 16; index = index + 1)
    fprintf(out, "%02X", get_wide_byte(head + index));
  fprintf(out, "\n");
  count = 0;
  for (x = get_first_cref_to(head); x != BADADDR;
       x = get_next_cref_to(head, x))
  {
    fprintf(out, "candidate_cref target_raw_ea=%08X from=%08X name=%s disasm=%s\n",
            ea, x, get_name(x), generate_disasm_line(x, 0));
    count = count + 1;
  }
  fprintf(out, "candidate_cref_count target_raw_ea=%08X count=%d\n", ea, count);
}

static scan_segment(out, seg)
{
  auto start;
  auto end;
  auto ea;
  auto word;
  auto name;

  start = get_segm_start(seg);
  end = get_segm_end(seg);
  name = get_segm_name(seg);
  fprintf(out, "scan_segment name=%s start=%08X end=%08X\n", name, start, end);
  ea = start;
  while (ea + 1 < end)
  {
    word = get_wide_word(ea);
    if (word == 0x2E10)
    {
      emit_candidate(out, ea, name, word);
      if (ea >= start + 8)
        emit_raw_window(out, ea - 8, ea + 10 < end ? ea + 10 : end);
    }
    ea = ea + 1;
  }
}

static main()
{
  auto out;
  auto seg;
  auto linear_ecl;

  auto_wait();
  out = fopen("/tmp/dos-start-2e10-candidates.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", get_input_file_path());
  fprintf(out, "basis=IDA database addresses plus raw segment offsets; target=ECL CALL 2E10h\n");
  fprintf(out, "semantic_status=unknown candidate only; no address-space alias is asserted\n");
  for (seg = get_first_seg(); seg != BADADDR; seg = get_next_seg(seg))
    scan_segment(out, seg);

  /* START.EXE is normally loaded at IDA base 10000h.  Keep both labels
   * explicit because 2E10h may be an ECL/segment offset rather than EA 2E10h. */
  linear_ecl = 0x10000 + 0x2E10;
  emit_xrefs(out, "ida_ea_00002E10", 0x2E10);
  emit_xrefs(out, "ida_ea_00012E10", linear_ecl);
  fclose(out);
  qexit(0);
}
