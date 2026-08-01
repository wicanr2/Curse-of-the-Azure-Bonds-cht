#include <idc.idc>

/*
 * Non-destructive IDA Pro 9.4 audit for PC-98 target visibility.
 *
 * Inputs are disposable code-only overlay copies.  Reports retain the
 * original overlay-local address and raw bytes; this script never renames
 * symbols or writes inferred semantics into the database.
 */

static emit_range(out, label, relative_start, relative_end)
{
  auto base, start, end, ea, size, index;
  base = get_inf_attr(INF_MIN_EA);
  start = base + relative_start;
  end = base + relative_end;
  if (end > get_inf_attr(INF_MAX_EA))
    end = get_inf_attr(INF_MAX_EA);
  del_items(start, DELIT_SIMPLE, end - start);
  ea = start;
  while (ea < end)
  {
    size = create_insn(ea);
    if (size <= 0)
      size = 1;
    fprintf(out, "label=%s ea=%08X local=%04X bytes=", label, ea, ea - base);
    for (index = 0; index < size; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(ea + index));
    fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
    ea = ea + size;
  }
}

static main()
{
  auto path, output_path, out, base, end, ea, size;
  auto_wait();
  path = get_input_file_path();
  base = get_inf_attr(INF_MIN_EA);
  end = get_inf_attr(INF_MAX_EA);
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(base, SEGATTR_BITNESS, 0);
  del_items(base, DELIT_SIMPLE, end - base);
  ea = base;
  while (ea < end)
  {
    size = create_insn(ea);
    if (size <= 0)
      size = 1;
    ea = ea + size;
  }

  if (strstr(path, "overlay-12.bin") != -1)
    output_path = "/work/pc98-visibility-overlay12.txt";
  else if (strstr(path, "overlay-13.bin") != -1)
    output_path = "/work/pc98-visibility-overlay13.txt";
  else
    qexit(1);

  out = fopen(output_path, "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s min=%08X max=%08X\n", path, base, end);

  if (strstr(path, "overlay-12.bin") != -1)
  {
    emit_range(out, "EFFECT_19_HANDLER", 0x06D0, 0x0780);
    emit_range(out, "EFFECT_47_HANDLER", 0x16E0, 0x17B0);
  }
  else
    emit_range(out, "CHECKTARGET", 0x11A0, 0x1260);

  fclose(out);
  qexit(0);
}
