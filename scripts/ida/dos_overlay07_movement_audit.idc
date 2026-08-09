#include <idc.idc>

/*
 * Non-destructive audit of the DOS overlay-07 routine around local 1B3F.
 * It records the exact wrap/update and vector-call bytes plus direct IDA
 * code xrefs in this overlay-local database.  It does not name the fields.
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

static emit_xrefs(out, target)
{
  auto from;
  auto count;

  count = 0;
  for (from = get_first_cref_to(target); from != BADADDR;
       from = get_next_cref_to(target, from))
  {
    fprintf(out, "direct_cref target=0x%04X from=0x%04X type=%d disasm=%s\n",
            target, from, XrefType(), generate_disasm_line(from, 0));
    count = count + 1;
  }
  fprintf(out, "direct_cref target=0x%04X count=%d\n", target, count);
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
  if (end <= 0x1BD6)
    qexit(2);
  decode_all(start, end);
  out = fopen("/tmp/dos-overlay07-movement-audit.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=overlay-07 local offset; direct IDA code xrefs only\n");
  fprintf(out, "semantic_status=unknown field names and movement-service meaning\n");
  emit_range(out, 0x1B3F, 0x1BD6);
  emit_xrefs(out, 0x1B3F);
  emit_xrefs(out, 0x1BB1);
  fclose(out);
  qexit(0);
}
