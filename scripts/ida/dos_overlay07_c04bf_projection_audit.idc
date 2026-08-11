#include <idc.idc>

/*
 * Non-destructive audit for the DOS ECL VM's C04Bh..C04Fh virtual-map
 * register dispatch.  Input must be a disposable extracted copy of GAME.OVR
 * overlay 07.  The report preserves overlay-local offsets and original bytes;
 * semantic labels are intentionally left to the reviewed Markdown ledger.
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
    fprintf(out, "local=0x%04X bytes=", ea);
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

  auto_wait();
  input = get_input_file_path();
  if (strstr(input, "overlay-07.bin") == -1)
    qexit(2);
  segment = get_first_seg();
  if (segment == BADADDR)
    qexit(2);
  set_segm_addressing(segment, 0);
  start = get_segm_start(segment);
  end = get_segm_end(segment);
  if (start != 0 || end <= 0x108B)
    qexit(2);

  decode_all(start, end);
  out = fopen("/tmp/dos-overlay07-c04bf-projection-audit.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=GAME.OVR extracted overlay-07 local offset\n");
  fprintf(out, "semantic_status=raw bytes and control flow only\n");
  fprintf(out, "segment_start=0x%04X segment_end=0x%04X\n", start, end);
  fprintf(out, "setter_window=0x0D70..0x0F38\n");
  emit_range(out, 0x0D70, 0x0F38);
  fprintf(out, "getter_window=0x0F3A..0x108B\n");
  emit_range(out, 0x0F3A, 0x108B);
  fclose(out);
  qexit(0);
}
