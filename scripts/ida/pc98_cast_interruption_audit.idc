#include <idc.idc>

/*
 * Non-destructive PC-98 casting-interruption audit.
 * The original overlays are mounted read-only; this script only emits an
 * additive instruction ledger with untouched local offsets and bytes.
 */
static emit_range(out, start_offset, end_offset)
{
  auto base, start, end, ea, size, index;
  base = get_inf_attr(INF_MIN_EA);
  start = base + start_offset;
  end = base + end_offset;
  if (end > get_inf_attr(INF_MAX_EA))
    end = get_inf_attr(INF_MAX_EA);
  del_items(start, DELIT_SIMPLE, end - start);
  ea = start;
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
  auto path, out, base, end;
  auto_wait();
  set_processor_type("8086", SETPROC_LOADER);
  base = get_inf_attr(INF_MIN_EA);
  end = get_inf_attr(INF_MAX_EA);
  set_segm_attr(base, SEGATTR_BITNESS, 0);
  path = get_input_file_path();
  if (strstr(path, "overlay-08.bin") != -1)
    out = fopen("/work/overlay08-full.txt", "w");
  else if (strstr(path, "overlay-23.bin") != -1)
    out = fopen("/work/overlay23-full.txt", "w");
  else if (strstr(path, "overlay-24.bin") != -1)
    out = fopen("/work/overlay24-full.txt", "w");
  else
    qexit(1);
  if (out == 0)
    qexit(2);
  emit_range(out, 0, end - base);
  fclose(out);
  qexit(0);
}
