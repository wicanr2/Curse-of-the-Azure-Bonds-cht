#include <idc.idc>

/*
 * Read-only address-space audit for overlay far calls written as 0A54:offset.
 * START.EXE's IDA image uses a +1000h paragraph load bias in this database:
 * runtime selector 0A54h is represented by IDA selector 1A54h.  The report
 * records that mapping, the original IDA symbol and candidate bytes without
 * renaming the database; the caller's buffer role remains an external inference.
 */

static decode_all(start, end)
{
  auto ea;
  auto size;

  del_items(start, DELIT_EXPAND, end - start);
  ea = start;
  while (ea < end)
  {
    size = create_insn(ea);
    if (size <= 0)
      size = 1;
    ea = ea + size;
  }
}

static emit_range(out, start, end)
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
    fprintf(out, "ea=0x%08X bytes=", ea);
    for (index = 0; index < size; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(ea + index));
    fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
    ea = ea + size;
  }
}

static emit_xrefs(out, target)
{
  auto from;
  auto count;

  count = 0;
  for (from = get_first_cref_to(target); from != BADADDR;
       from = get_next_cref_to(target, from))
  {
    fprintf(out, "direct_cref target=0x%08X from=0x%08X type=%d disasm=%s\n",
            target, from, XrefType(), generate_disasm_line(from, 0));
    count = count + 1;
  }
  fprintf(out, "direct_cref target=0x%08X count=%d\n", target, count);
}

static main()
{
  auto input;
  auto seg;
  auto out;
  auto target;
  auto target_seg;

  auto_wait();
  input = get_input_file_path();
  if (strstr(input, "START.EXE") == -1)
    qexit(2);
  target = 0x1A540 + 0x0329;
  target_seg = BADADDR;
  for (seg = get_first_seg(); seg != BADADDR; seg = get_next_seg(seg))
  {
    if (get_segm_start(seg) <= target && target < get_segm_end(seg))
    {
      target_seg = seg;
      break;
    }
  }
  if (target_seg == BADADDR)
    qexit(2);
  decode_all(get_segm_start(target_seg), get_segm_end(target_seg));
  out = fopen("/tmp/dos-start-0a54-call-audit.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=runtime 0A54:0329 candidate mapped to IDA +1000h paragraph bias\n");
  fprintf(out, "semantic_status=exact target mapping／original IDA symbol; caller buffer role=strong inference\n");
  fprintf(out, "ida_name=%s\n", get_name(target));
  fprintf(out, "runtime_far=0A54:0329 ida_ea=0x%08X ida_selector=0x1A54\n", target);
  fprintf(out, "segment name=%s start=0x%08X end=0x%08X sel=0x%04X\n",
          get_segm_name(target_seg), get_segm_start(target_seg),
          get_segm_end(target_seg), get_segm_attr(target_seg, SEGATTR_SEL));
  fprintf(out, "-- target context --\n");
  emit_range(out, target - 0x30, target + 0x90);
  emit_xrefs(out, target);
  fclose(out);
  qexit(0);
}
