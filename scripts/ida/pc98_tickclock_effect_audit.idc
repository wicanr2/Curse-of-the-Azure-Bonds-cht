#include <idc.idc>

/*
 * Non-destructive IDA Pro 9.4 audit for PC-98 CLOCK_ overlay 20.
 * Run on a writable copy of the code-only overlay.  Original executable,
 * overlay container, extracted bytes and baseline databases remain read-only.
 */

static emit_range(out, label, relative_start, relative_end)
{
  auto base;
  auto start;
  auto end;
  auto ea;
  auto size;
  auto index;

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
  auto path;
  auto out;
  auto base;

  auto_wait();
  path = get_input_file_path();
  if (strstr(path, "overlay-20.bin") == -1)
    qexit(1);
  out = fopen("/work/pc98-tickclock-effect-overlay20.txt", "w");
  if (out == 0)
    qexit(2);

  base = get_inf_attr(INF_MIN_EA);
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(base, SEGATTR_BITNESS, 0);
  fprintf(out, "input=%s min=%08X max=%08X\n", path, base,
          get_inf_attr(INF_MAX_EA));

  /* Borland CLOCK_ symbols: TICKCLOCK 03B9h, REST 0CD0h. */
  emit_range(out, "CLOCK_PREFIX", 0x0000, 0x03B9);
  emit_range(out, "TICKCLOCK", 0x03B9, 0x0CD0);

  fclose(out);
  qexit(0);
}
