#include <idc.idc>

/*
 * Non-destructive PC-98 Sleep spell audit.
 *
 * Run only against a disposable copy of overlay-22.bin.  The script appends
 * scope labels to a text ledger while preserving original offsets, bytes and
 * disassembly; it never renames symbols or patches the source image.
 */
static emit_range(out, label, start, end)
{
  auto ea, size, index;
  fprintf(out, "[%s] range=%04X..%04X confidence=exact-bytes\n", label, start, end);
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
  if (strstr(input, "overlay-22.bin") == -1)
    qexit(1);
  output = "/work/overlay-22-sleep-ledger.txt";
  out = fopen(output, "w");
  if (out == 0)
    qexit(2);
  emit_range(out, "EFFECT_DURATION_HELPER", 0x0E75, 0x0F62);
  emit_range(out, "COMMON_TARGET_EFFECT_WRITER", 0x0F62, 0x1119);
  emit_range(out, "SLEEP_HANDLER_ENTRY_41", 0x2547, 0x267D);
  emit_range(out, "SLEEP_DISPATCH_SLOT_15", 0x6816, 0x6823);
  fclose(out);
  qexit(0);
}
