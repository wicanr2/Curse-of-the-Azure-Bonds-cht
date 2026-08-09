#include <idc.idc>

/*
 * Read-only audit of resident targets used by overlay-30's four-plane copy.
 * Runtime far selectors are kept separate from IDA linear EAs.  The script
 * preserves original IDA names and emits raw bytes, direct xrefs and the
 * address-space mapping; it does not rename the database or infer map rules.
 */

static decode_range(start, end)
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

  fprintf(out, "-- context 0x%08X..0x%08X --\n", start, end);
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
  auto index;

  count = 0;
  for (from = get_first_cref_to(target); from != BADADDR;
       from = get_next_cref_to(target, from))
  {
    fprintf(out, "direct_cref target=0x%08X from=0x%08X type=%d name=%s disasm=%s\n",
            target, from, XrefType(), get_name(from), generate_disasm_line(from, 0));
    fprintf(out, "caller_bytes from=0x%08X bytes=", from);
    for (index = 0; index < 8; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(from + index));
    fprintf(out, "\n");
    count = count + 1;
  }
  fprintf(out, "direct_cref target=0x%08X count=%d\n", target, count);
}

static emit_function(out, target)
{
  auto start;
  auto end;

  start = get_func_attr(target, FUNCATTR_START);
  end = get_func_attr(target, FUNCATTR_END);
  if (start == BADADDR || end == BADADDR || start == 0 || end <= start)
  {
    fprintf(out, "function target=0x%08X status=unavailable\n", target);
    return;
  }
  fprintf(out, "function target=0x%08X start=0x%08X end=0x%08X\n",
          target, start, end);
  emit_range(out, start, end);
}

static emit_target(out, runtime_offset)
{
  auto target;
  auto segment;
  auto start;
  auto end;

  target = 0x1A540 + runtime_offset;
  segment = BADADDR;
  for (segment = get_first_seg(); segment != BADADDR;
       segment = get_next_seg(segment))
  {
    if (get_segm_start(segment) <= target && target < get_segm_end(segment))
      break;
  }
  if (segment == BADADDR)
  {
    fprintf(out, "target runtime 0A54:%04X ida_ea=0x%08X status=unmapped\n",
            runtime_offset, target);
    return;
  }
  start = get_segm_start(segment);
  end = get_segm_end(segment);
  fprintf(out, "target runtime 0A54:%04X ida_ea=0x%08X selector=0x%04X segment=%s\n",
          runtime_offset, target, get_segm_attr(segment, SEGATTR_SEL),
          get_segm_name(segment));
  fprintf(out, "ida_name=%s\n", get_name(target));
  emit_range(out, target - 0x40, target + 0x500);
  emit_xrefs(out, target);
}

static emit_target_at(out, runtime_far, target, selector)
{
  auto segment;

  segment = BADADDR;
  for (segment = get_first_seg(); segment != BADADDR;
       segment = get_next_seg(segment))
  {
    if (get_segm_start(segment) <= target && target < get_segm_end(segment))
      break;
  }
  if (segment == BADADDR)
  {
    fprintf(out, "target runtime %s ida_ea=0x%08X status=unmapped\n",
            runtime_far, target);
    return;
  }
  fprintf(out, "target runtime %s ida_ea=0x%08X selector=0x%04X segment=%s\n",
          runtime_far, target, selector, get_segm_name(segment));
  fprintf(out, "ida_name=%s\n", get_name(target));
  emit_function(out, target);
  emit_range(out, target - 0x40, target + 0x100);
  emit_xrefs(out, target);
}

static main()
{
  auto input;
  auto out;
  auto segment;

  auto_wait();
  input = get_input_file_path();
  if (strstr(input, "START.EXE") == -1)
    qexit(2);
  segment = get_first_seg();
  if (segment == BADADDR)
    qexit(2);
  decode_range(get_segm_start(segment), get_segm_end(segment));
  out = fopen("/tmp/dos-start-buffer-routines.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=runtime selector 0A54h mapped to IDA +1000h paragraph bias\n");
  fprintf(out, "semantic_status=exact target mapping／original IDA symbols／raw callers; buffer copy role remains evidence-scoped\n");
  emit_target(out, 0x1ABD);
  emit_target(out, 0x0364);
  emit_target(out, 0x0634);
  emit_target(out, 0x06C1);
  emit_target_at(out, "0636:08DE", 0x16360 + 0x08DE, 0x1636);
  fclose(out);
  qexit(0);
}
