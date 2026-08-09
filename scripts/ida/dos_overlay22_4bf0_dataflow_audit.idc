#include <idc.idc>

/*
 * Non-destructive, bounded IDA report for the DOS overlay-22 candidate
 * work-cell flow.  The input must be a disposable code-only overlay-22 copy;
 * no names, comments, or types are written to the source binary or baseline
 * database.  The report deliberately preserves overlay-local offsets and
 * labels every 4BF0h hit as a candidate until a resident writer/consumer is
 * closed by runtime evidence.
 */

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

static main()
{
  auto input;
  auto segment;
  auto base;
  auto end;
  auto ea;
  auto word;
  auto out;
  auto hit_count;

  input = get_input_file_path();
  if (strstr(input, "overlay-22.bin") == -1)
    qexit(2);
  segment = get_first_seg();
  if (segment == BADADDR)
    qexit(2);
  set_segm_addressing(segment, 0);
  base = get_segm_start(segment);
  end = get_segm_end(segment);
  if (end < 0x0C00)
    qexit(2);

  /* Only the disposable copy is decoded, and only around the known flow. */
  decode_range(0x0300, 0x04C0);
  decode_range(0x0900, 0x0C00);

  out = fopen("/tmp/dos-overlay22-4bf0-dataflow.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=overlay-local code offset; DOS GAME.OVR overlay 22\n");
  fprintf(out, "semantic_status=unknown candidate writer/reader flow only\n");
  fprintf(out, "candidate_reads=overlay-local 0x03CF..0x03D5\n");
  fprintf(out, "candidate_writes=overlay-local 0x0969..0x0970 and 0x099D..0x09A3\n");
  fprintf(out, "-- candidate flow --\n");
  emit_range(out, 0x0300, 0x03F0);
  emit_range(out, 0x03F5, 0x04C0);
  emit_range(out, 0x0900, 0x0C00);
  fprintf(out, "-- raw candidate offsets --\n");
  hit_count = 0;
  ea = base;
  while (ea + 1 < end)
  {
    word = get_wide_word(ea);
    if (word == 0x4BF0 || word == 0x4BF1 || word == 0x4BF2 || word == 0x4BF3)
    {
      fprintf(out, "RAW_LE16 local=0x%04X target=0x%04X bytes=%02X%02X disasm=%s\n",
              ea - base, word & 0xFFFF, get_wide_byte(ea),
              get_wide_byte(ea + 1), generate_disasm_line(ea, 0));
      hit_count = hit_count + 1;
    }
    ea = ea + 1;
  }
  fprintf(out, "RAW_LE16_COUNT=%d\n", hit_count);
  fclose(out);
  qexit(0);
}
