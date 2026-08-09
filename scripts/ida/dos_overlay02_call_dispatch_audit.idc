#include <idc.idc>

/*
 * Non-destructive, bounded audit for DOS GAME.OVR extracted overlay 02.  The
 * public CoAB source labels the corresponding command dispatcher as ovr003;
 * the extraction used by this project is zero-based, so the input filename
 * and overlay-local address remain the authority until that numbering bridge
 * is independently proven.
 *
 * The input is a disposable code-only overlay copy. Numeric hits are reported
 * with overlay-local offsets and remain candidates; this script does not
 * rename functions or infer external routine semantics.
 */

static is_target(word)
{
  return word == 0x7FFF || word == 0x2E10 || word == 0xC01E ||
         word == 0xB200 || word == 0xAE11 || word == 0x401F ||
         word == 0x3201;
}

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

static emit_window(out, ea, start, end)
{
  auto index;
  auto left;
  auto right;

  left = ea > start + 16 ? ea - 16 : start;
  right = ea + 24 < end ? ea + 24 : end;
  fprintf(out, "raw_window local=0x%04X bytes=", left);
  for (index = left; index < right; index = index + 1)
    fprintf(out, "%02X", get_wide_byte(index));
  fprintf(out, "\n");
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
  auto ea;
  auto word;
  auto head;
  auto size;
  auto out;
  auto hit_count;

  auto_wait();
  input = get_input_file_path();
  if (strstr(input, "overlay-02.bin") == -1)
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
  out = fopen("/tmp/dos-overlay02-call-dispatch.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=GAME.OVR overlay-local code offset; extracted overlay 02\n");
  fprintf(out, "reference_label=public CoAB source ovr003; extraction numbering bridge not asserted\n");
  fprintf(out, "semantic_status=unknown candidate dispatch constants only\n");
  fprintf(out, "targets=7FFF,2E10,C01E,B200,AE11,401F,3201\n");
  fprintf(out, "segment_start=0x%04X segment_end=0x%04X\n", start, end);
  hit_count = 0;
  ea = start;
  while (ea + 1 < end)
  {
    word = get_wide_word(ea);
    if (is_target(word))
    {
      head = get_item_head(ea);
      size = get_item_size(head);
      if (size <= 0)
        size = 1;
      fprintf(out, "candidate local=0x%04X target=0x%04X item_head=0x%04X item_size=%d code=%d name=%s disasm=%s bytes=",
              ea - start, word, head, size, is_code(get_flags(head)),
              get_name(head), generate_disasm_line(head, 0));
      fprintf(out, "%02X%02X\n", get_wide_byte(ea), get_wide_byte(ea + 1));
      emit_window(out, ea, start, end);
      hit_count = hit_count + 1;
    }
    ea = ea + 1;
  }
  fprintf(out, "candidate_count=%d\n", hit_count);
  fprintf(out, "-- dispatch candidate window --\n");
  emit_range(out, 0x2EA0, 0x30C0);
  fclose(out);
  qexit(0);
}
