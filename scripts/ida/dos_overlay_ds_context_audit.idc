#include <idc.idc>

/*
 * Non-destructive context export for direct DS-field candidates.
 *
 * The ranges are deliberately limited to callers/readers/writers found by
 * dos_overlay_ds_field_audit.idc.  They preserve overlay-local addresses and
 * continuous instructions; no semantic names or cross-overlay address joins
 * are created here.
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
  auto start;
  auto end;
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

  out = fopen("/tmp/dos-overlay-ds-context.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=continuous decoded overlay-local instructions\n");
  fprintf(out, "semantic_status=unknown; ranges selected from direct DS-field candidates\n");

  if (strstr(input, "overlay-02.bin") != -1)
  {
    emit_range(out, 0x03D0, 0x0460);
    emit_range(out, 0x2EF0, 0x3060);
    emit_range(out, 0x3690, 0x3710);
  }
  else if (strstr(input, "overlay-07.bin") != -1)
  {
    emit_range(out, 0x0E80, 0x0F20);
    emit_range(out, 0x1B20, 0x1BE0);
  }
  else if (strstr(input, "overlay-11.bin") != -1)
  {
    emit_range(out, 0x00C0, 0x0460);
    emit_range(out, 0x0750, 0x0810);
  }
  else if (strstr(input, "overlay-14.bin") != -1)
  {
    emit_range(out, 0x003E, 0x02A0);
    emit_range(out, 0x0780, 0x0AE0);
  }
  else if (strstr(input, "overlay-28.bin") != -1)
  {
    emit_range(out, 0x0000, 0x01C0);
  }
  else if (strstr(input, "overlay-30.bin") != -1)
  {
    emit_range(out, 0x02A0, 0x0410);
    emit_range(out, 0x0580, 0x0800);
    emit_range(out, 0x11C0, 0x1320);
    emit_range(out, 0x13C0, 0x147F);
  }
  else
  {
    fprintf(out, "no configured context range for this overlay\n");
  }
  fclose(out);
  qexit(0);
}
