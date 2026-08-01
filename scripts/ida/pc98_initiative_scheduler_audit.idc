#include <idc.idc>

/* Non-destructive PC-98 initiative/action-delay scheduler audit. */

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

  if (strstr(path, "overlay-08.bin") != -1)
    output_path = "/work/pc98-initiative-overlay08.txt";
  else if (strstr(path, "overlay-09.bin") != -1)
    output_path = "/work/pc98-initiative-overlay09.txt";
  else if (strstr(path, "overlay-13.bin") != -1)
    output_path = "/work/pc98-initiative-overlay13.txt";
  else if (strstr(path, "overlay-14.bin") != -1)
    output_path = "/work/pc98-initiative-overlay14.txt";
  else if (strstr(path, "overlay-24.bin") != -1)
    output_path = "/work/pc98-initiative-overlay24.txt";
  else
    qexit(1);

  out = fopen(output_path, "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s min=%08X max=%08X\n", path, base, end);

  if (strstr(path, "overlay-08.bin") != -1)
  {
    emit_range(out, "TURN_SCHEDULER_CANDIDATE", 0x0180, 0x0480);
    emit_range(out, "ACTION_DELAY_COMPLETION", 0x0580, 0x0720);
  }
  else if (strstr(path, "overlay-09.bin") != -1)
    emit_range(out, "TURN_SCHEDULER_D100_CANDIDATE", 0x09C0, 0x0C20);
  else if (strstr(path, "overlay-13.bin") != -1)
    emit_range(out, "INITIATIVE_CALLER", 0x0000, 0x0200);
  else if (strstr(path, "overlay-14.bin") != -1)
    emit_range(out, "INITIATIVE_CANDIDATES", 0x0000, 0x0800);
  else
    emit_range(out, "DEXRABONUS", 0x13D0, 0x14B0);

  fclose(out);
  qexit(0);
}
