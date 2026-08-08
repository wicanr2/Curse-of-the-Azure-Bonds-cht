#include <idc.idc>

/*
 * Non-destructive IDA Pro 9.4 audit for PC-98 callers of the shared
 * target-list handler.  This script only runs on copied code-only overlays;
 * it preserves overlay-local addresses and raw instruction bytes and never
 * renames functions or writes semantic types into a baseline database.
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
    fprintf(out, "label=%s ea=%08X local=%04X bytes=", label, ea,
            ea - base);
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
  if (strstr(path, "overlay-09.bin") != -1)
    output_path = "/work/pc98-general-target-overlay09.txt";
  else if (strstr(path, "overlay-13.bin") != -1)
    output_path = "/work/pc98-general-target-overlay13.txt";
  else
    qexit(1);

  out = fopen(output_path, "w");
  if (out == 0)
    qexit(2);
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
  fprintf(out, "input=%s min=%08X max=%08X\n", path, base, end);

  if (strstr(path, "overlay-09.bin") != -1)
  {
    emit_range(out, "CALLER_0430", 0x0380, 0x0500);
    emit_range(out, "CALLER_1010", 0x0F70, 0x1180);
    emit_range(out, "LOCAL_119F_AND_NEIGHBORS", 0x1180, 0x13A0);
    emit_range(out, "CALLER_19C0", 0x18C0, 0x1A80);
  }
  else
  {
    emit_range(out, "CALLER_2F40", 0x2F40, 0x3090);
    emit_range(out, "CALLER_34F0", 0x34F0, 0x36A0);
    emit_range(out, "PICKTARGET_CALLSITE", 0x3DE0, 0x3EB0);
  }
  fclose(out);
  qexit(0);
}
