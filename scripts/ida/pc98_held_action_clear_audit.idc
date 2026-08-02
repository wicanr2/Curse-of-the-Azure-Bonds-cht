#include <idc.idc>

/*
 * Non-destructive PC-98 held-effect action-clear audit.
 *
 * Source overlays are mounted read-only and copied into a disposable /work
 * directory.  The report preserves overlay-local offsets, bytes and original
 * disassembly.  It never renames symbols or patches source bytes.
 */
static emit_range(out, label, start, end)
{
  auto ea, size, index;
  fprintf(out, "[%s] range=%04X..%04X\n", label, start, end);
  del_items(start, DELIT_SIMPLE, end - start);
  ea = start;
  while (ea < end)
  {
    size = create_insn(ea);
    if (size <= 0)
      size = 1;
    fprintf(out, "local=%04X bytes=", ea);
    for (index = 0; index < size; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(ea + index));
    fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
    ea = ea + size;
  }
}

static main()
{
  auto input, output, out, base;
  auto_wait();
  set_processor_type("8086", SETPROC_LOADER);
  base = get_inf_attr(INF_MIN_EA);
  set_segm_attr(base, SEGATTR_BITNESS, 0);
  input = get_input_file_path();
  if (strstr(input, "overlay-12.bin") != -1)
    output = "/work/overlay-12-held-ledger.txt";
  else if (strstr(input, "overlay-24.bin") != -1)
    output = "/work/overlay-24-held-ledger.txt";
  else
    qexit(1);
  out = fopen(output, "w");
  if (out == 0)
    qexit(2);
  if (strstr(input, "overlay-12.bin") != -1)
  {
    emit_range(out, "HELD_EFFECT_HANDLER", 0x0075, 0x008B);
    emit_range(out, "EFFECT_1F_TABLE_SLOT", 0x305D, 0x306A);
    emit_range(out, "EFFECT_33_35_TABLE_SLOTS", 0x3161, 0x3188);
  }
  else
    emit_range(out, "CLEARACTION_CONSUMER", 0x2A5B, 0x2A9F);
  fclose(out);
  qexit(0);
}
