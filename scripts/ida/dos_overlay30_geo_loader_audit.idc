#include <idc.idc>

/*
 * Non-destructive audit of overlay-30's GEO resource-name builder and the
 * decoded-size gate immediately before the DS:7206h four-plane copies.
 * Keep the raw data bytes, overlay-local offsets, and decoded instructions
 * side by side; this script does not rename or comment the IDA database.
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
  fprintf(out, "\n");
}

static emit_data_bytes(out, start, end)
{
  auto ea;

  fprintf(out, "raw_data local=0x%04X..0x%04X bytes=", start, end - 1);
  emit_bytes(out, start, end);
}

static emit_code(out, start, end)
{
  auto ea;
  auto size;
  auto line;

  ea = start;
  while (ea < end)
  {
    size = get_item_size(ea);
    if (size <= 0)
      size = 1;
    line = generate_disasm_line(ea, 0);
    fprintf(out, "code local=0x%04X bytes=", ea);
    emit_bytes(out, ea, ea + size);
    fprintf(out, " disasm=%s\n", line);
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
  if (strstr(input, "overlay-30.bin") == -1)
    qexit(2);
  segment = get_first_seg();
  if (segment == BADADDR)
    qexit(2);
  set_segm_addressing(segment, 0);
  start = get_segm_start(segment);
  end = get_segm_end(segment);

  decode_all(start, end);
  out = fopen("/tmp/dos-overlay30-geo-loader.txt", "w");
  if (out == 0)
    qexit(2);

  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=overlay-30 local offset; raw data plus IDA decoded code\n");
  fprintf(out, "semantic_status=source filename／decoded-size gate evidence only; no generic rename\n");
  fprintf(out, "resource_prefix_data_local=0x1310\n");
  emit_data_bytes(out, 0x1310, 0x133A);
  fprintf(out, "resource_prefix_interpretation=raw Pascal bytes 03 'GEO' and 04 '.dax'; remainder is adjacent message data\n");
  fprintf(out, "resource_area_source=DS:5BEEh read at local 0x1341; width word 1 before resident 0A54:12ABh call\n");
  fprintf(out, "resource_loader_call=local 0x1393 far 0636:08DEh; output word [bp-2], output pointer [bp-6], selector [bp+6]\n");
  fprintf(out, "decoded_size_gate=local 0x139E cmp word [bp-2],0402h; success branch 0x13D3\n");
  fprintf(out, "failure_gate=local 0x1398 zero result or non-0402h result takes error／cleanup path\n");
  fprintf(out, "copy_source_base=output pointer + 0x002h\n");
  fprintf(out, "copy_destinations=DS:7206h + 0x000／0x100／0x200／0x300; each count 0x100h\n");
  fprintf(out, "copy_free=resident 0A54:0364h with same output pointer and returned word\n");
  fprintf(out, "-- resource builder／gate code --\n");
  emit_code(out, 0x1341, 0x13D3);
  fprintf(out, "-- return／copy handoff code --\n");
  emit_code(out, 0x13D3, 0x1476);
  fclose(out);
  qexit(0);
}
