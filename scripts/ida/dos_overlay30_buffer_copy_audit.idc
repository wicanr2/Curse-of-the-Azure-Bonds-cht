#include <idc.idc>

/*
 * Non-destructive continuous dump of the overlay-30 routine that copies the
 * four 0x100-byte DS:7206 planes through Borland Move.  It keeps the
 * overlay-local address and raw bytes together; semantic names remain in the
 * external evidence ledger.
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

  fprintf(out, "-- range 0x%04X..0x%04X --\n", start, end);
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
  auto out;

  auto_wait();
  input = get_input_file_path();
  if (strstr(input, "overlay-30.bin") == -1)
    qexit(2);
  segment = get_first_seg();
  if (segment == BADADDR)
    qexit(2);
  set_segm_addressing(segment, 0);
  decode_all(get_segm_start(segment), get_segm_end(segment));
  out = fopen("/tmp/dos-overlay30-buffer-copy.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=overlay-30 local offset; continuous IDA decode\n");
  fprintf(out, "semantic_status=exact copy call shape／plane offsets; destination role remains evidence-scoped\n");
  emit_range(out, 0x1300, 0x1480);
  fclose(out);
  qexit(0);
}
