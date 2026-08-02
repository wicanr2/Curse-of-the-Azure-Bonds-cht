#include <idc.idc>

/*
 * Non-destructive PC-98 OBJECTLIST audit. The input overlay remains read-only;
 * semantic labels exist only in this report and carry no database rename.
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

  if (strstr(path, "overlay-10.bin") != -1)
    output_path = "/work/pc98-objectlist-overlay10.txt";
  else if (strstr(path, "overlay-13.bin") != -1)
    output_path = "/work/pc98-objectlist-overlay13.txt";
  else if (strstr(path, "overlay-23.bin") != -1)
    output_path = "/work/pc98-objectlist-overlay23.txt";
  else if (strstr(path, "overlay-32.bin") != -1)
    output_path = "/work/pc98-objectlist-overlay32.txt";
  else
    qexit(1);

  out = fopen(output_path, "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s min=%08X max=%08X\n", path, base, end);
  if (strstr(path, "overlay-10.bin") != -1)
  {
    emit_range(out, "OBJECTLIST_COORD_WRITER", 0x1200, 0x1330);
    emit_range(out, "OBJECTLIST_BUILDER", 0x17C0, 0x1C20);
  }
  else if (strstr(path, "overlay-13.bin") != -1)
  {
    emit_range(out, "OBJECTLIST_COORD_UPDATE", 0x0700, 0x0B00);
    emit_range(out, "OBJECTLIST_CLEAR", 0x4860, 0x4930);
  }
  else if (strstr(path, "overlay-23.bin") != -1)
    emit_range(out, "OBJECTLIST_CLEAR", 0x14D0, 0x1570);
  else
    emit_range(out, "OBJECTLIST_QUERY_AND_WRITER", 0x1280, 0x1A80);
  fclose(out);
  qexit(0);
}
