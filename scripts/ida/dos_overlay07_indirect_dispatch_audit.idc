#include <idc.idc>

/*
 * Non-destructive audit of possible indirect entry paths to overlay-07's
 * local 1B3F routine.  Direct IDA xrefs alone cannot see a pointer/table
 * dispatch, so this report lists only decoded FF-form indirect call/jump
 * instructions and raw LE16 candidates without naming their target semantics.
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

static emit_indirect(out, start, end)
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
    if (is_code(get_flags(ea)))
    {
      line = generate_disasm_line(ea, 0);
      /* FF /2 and FF /3 are the 16-bit near/far indirect call forms;
       * FF /4 and FF /5 are the corresponding indirect jumps.  Do not
       * classify every far call containing "ptr" as an indirect candidate:
       * the control-vector evidence is audited separately. */
      if (get_wide_byte(ea) == 0xFF &&
          (strstr(line, "call") != -1 || strstr(line, "jmp") != -1))
      {
        if (strstr(line, "]") != -1)
        {
          fprintf(out, "indirect_FF_call_or_jump local=0x%04X bytes=",
                  ea - start);
          emit_bytes(out, ea, size);
          fprintf(out, " disasm=%s\n", line);
        }
      }
    }
    ea = ea + size;
  }
}

static emit_raw_candidates(out, start, end)
{
  auto ea;
  auto word;
  auto count;

  count = 0;
  ea = start;
  while (ea + 1 < end)
  {
    word = get_wide_word(ea);
    if ((word & 0xFFFF) == 0x1B3F)
    {
      fprintf(out, "raw_LE16_target_1B3F local=0x%04X bytes=",
              ea - start);
      emit_bytes(out, ea, 2);
      fprintf(out, "\n");
      count = count + 1;
    }
    ea = ea + 1;
  }
  fprintf(out, "raw_LE16_target_1B3F_count=%d\n", count);
}

static main()
{
  auto input;
  auto segment;
  auto start;
  auto end;
  auto out;
  auto from;
  auto count;

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
  decode_all(start, end);
  out = fopen("/tmp/dos-overlay07-indirect-dispatch.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=overlay-07 local offset; IDA decoded code plus raw LE16 scan\n");
  fprintf(out, "semantic_status=unknown dispatch target; no pointer/table hit is promoted\n");
  fprintf(out, "target_local=0x1B3F direct_cref_count=");
  count = 0;
  for (from = get_first_cref_to(0x1B3F); from != BADADDR;
       from = get_next_cref_to(0x1B3F, from))
    count = count + 1;
  fprintf(out, "%d\n", count);
  emit_indirect(out, start, end);
  emit_raw_candidates(out, start, end);
  fclose(out);
  qexit(0);
}
