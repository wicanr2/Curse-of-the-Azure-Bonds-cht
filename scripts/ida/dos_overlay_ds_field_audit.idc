#include <idc.idc>

/*
 * Non-destructive candidate audit for DS-relative fields used by DOS overlays.
 *
 * Each extracted overlay is decoded in its own disposable IDA database.  The
 * report keeps the overlay-local offset, raw instruction bytes, and IDA
 * disassembly together.  A hit is only a direct decoded-operand candidate:
 * this script does not infer a field name, caller/consumer relation, or map
 * semantics from a matching hexadecimal number.
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

static emit_bytes(out, ea, size)
{
  auto index;

  for (index = 0; index < size; index = index + 1)
    fprintf(out, "%02X", get_wide_byte(ea + index));
}

static emit_if_match(out, ea, size, line, field)
{
  if (strstr(line, field) == -1)
    return;
  fprintf(out, "candidate field=%s local=0x%04X bytes=", field, ea);
  emit_bytes(out, ea, size);
  fprintf(out, " disasm=%s\n", line);
}

static main()
{
  auto input;
  auto segment;
  auto start;
  auto end;
  auto ea;
  auto size;
  auto line;
  auto out;

  auto_wait();
  input = get_input_file_path();
  if (strstr(input, "overlay-") == -1)
    qexit(2);
  segment = get_first_seg();
  if (segment == BADADDR)
    qexit(2);
  set_segm_addressing(segment, 0);
  start = get_segm_start(segment);
  end = get_segm_end(segment);
  if (end <= start)
    qexit(2);

  decode_all(start, end);
  out = fopen("/tmp/dos-overlay-ds-fields.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=decoded overlay-local instructions; direct DS operand text only\n");
  fprintf(out, "semantic_status=unknown field meaning; candidate hits are not consumers\n");
  fprintf(out, "segment_start=0x%04X segment_end=0x%04X\n", start, end);

  ea = start;
  while (ea < end)
  {
    size = get_item_size(ea);
    if (size <= 0)
      size = 1;
    if (is_code(get_flags(ea)))
    {
      line = generate_disasm_line(ea, 0);
      emit_if_match(out, ea, size, line, "7206h");
      emit_if_match(out, ea, size, line, "720Fh");
      emit_if_match(out, ea, size, line, "7210h");
      emit_if_match(out, ea, size, line, "7211h");
      emit_if_match(out, ea, size, line, "7212h");
      emit_if_match(out, ea, size, line, "7213h");
      emit_if_match(out, ea, size, line, "8B5Eh");
    }
    ea = ea + size;
  }
  fclose(out);
  qexit(0);
}
