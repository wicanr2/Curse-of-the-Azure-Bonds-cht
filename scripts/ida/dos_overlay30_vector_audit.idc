#include <idc.idc>

/*
 * Non-destructive, bounded IDA audit for the DOS GAME.OVR overlay-30 routine
 * selected by START.EXE control vector 017F:003E.
 *
 * The input is an extracted overlay copy.  The vector/control-block mapping is
 * recorded by the surrounding spec; this script only decodes and exports the
 * overlay-local code range so the original bytes, address space, and IDA
 * disassembly remain independently reviewable.
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

static emit_bytes(out, start, end)
{
  auto ea;

  for (ea = start; ea < end; ea = ea + 1)
    fprintf(out, "%02X", get_wide_byte(ea));
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
    fprintf(out, "range local=0x%04X bytes=", ea);
    for (index = 0; index < size; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(ea + index));
    fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
    ea = ea + size;
  }
}

static main()
{
  auto input;
  auto segment;
  auto start;
  auto end;
  auto out;
  auto target;
  auto target_size;

  auto_wait();
  input = get_input_file_path();
  if (strstr(input, "overlay-30.bin") == -1)
    qexit(2);
  segment = get_first_seg();
  if (segment == BADADDR)
    qexit(2);
  set_segm_addressing(segment, 0);
  start = get_segm_start(segment);
  end = get_segm_end(segment);
  if (end <= start || start > 0x07C6 || end <= 0x07C6)
    qexit(2);

  decode_all(start, end);
  out = fopen("/tmp/dos-overlay30-vector-audit.txt", "w");
  if (out == 0)
    qexit(2);

  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=GAME.OVR extracted overlay-30 local offset\n");
  fprintf(out, "semantic_status=unknown external routine semantics; exact bytes/control flow only\n");
  fprintf(out, "requested_vector=START control vector 017F:003E\n");
  fprintf(out, "requested_vector_index=6 zero-based; target_local=0x07C6\n");
  fprintf(out, "segment_start=0x%04X segment_end=0x%04X\n", start, end);

  target = 0x07C6;
  target_size = get_item_size(target);
  if (target_size <= 0)
    target_size = 1;
  fprintf(out, "target local=0x%04X item_size=%d code=%d name=%s disasm=%s bytes=",
          target, target_size, is_code(get_flags(target)), get_name(target),
          generate_disasm_line(target, 0));
  emit_bytes(out, target, target + target_size);
  fprintf(out, "\n");
  fprintf(out, "-- continuous target window --\n");
  emit_range(out, 0x07C0, 0x1020);
  fclose(out);
  qexit(0);
}
