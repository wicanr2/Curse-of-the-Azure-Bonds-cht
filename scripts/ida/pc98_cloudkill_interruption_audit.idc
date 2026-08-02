#include <idc.idc>

/*
 * Non-destructive PC-98 Cloudkill/interruption audit.
 *
 * Inputs are copied from read-only original overlays into a disposable IDA
 * workspace.  This script emits an additive ledger containing untouched
 * overlay-local offsets, bytes, and disassembly; it never renames or patches
 * the source database.
 */
static emit_all(out)
{
  auto base, end, ea, size, index;
  base = get_inf_attr(INF_MIN_EA);
  end = get_inf_attr(INF_MAX_EA);
  del_items(base, DELIT_SIMPLE, end - base);
  ea = base;
  while (ea < end)
  {
    size = create_insn(ea);
    if (size <= 0)
      size = 1;
    fprintf(out, "ea=%08X local=%04X bytes=", ea, ea - base);
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
  if (strstr(input, "overlay-02.bin") != -1)
    output = "/work/overlay-02-ledger.txt";
  else if (strstr(input, "overlay-07.bin") != -1)
    output = "/work/overlay-07-ledger.txt";
  else if (strstr(input, "overlay-12.bin") != -1)
    output = "/work/overlay-12-ledger.txt";
  else if (strstr(input, "overlay-13.bin") != -1)
    output = "/work/overlay-13-ledger.txt";
  else if (strstr(input, "overlay-19.bin") != -1)
    output = "/work/overlay-19-ledger.txt";
  else if (strstr(input, "overlay-22.bin") != -1)
    output = "/work/overlay-22-ledger.txt";
  else if (strstr(input, "overlay-23.bin") != -1)
    output = "/work/overlay-23-ledger.txt";
  else if (strstr(input, "overlay-24.bin") != -1)
    output = "/work/overlay-24-ledger.txt";
  else
    qexit(1);
  out = fopen(output, "w");
  if (out == 0)
    qexit(2);
  emit_all(out);
  fclose(out);
  qexit(0);
}
