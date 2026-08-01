#include <idc.idc>

static emit_range(out, label, relative_start, relative_end)
{
  auto base, start, end, ea, size, index;
  base = get_inf_attr(INF_MIN_EA);
  start = base + relative_start;
  end = base + relative_end;
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
  auto out, ea, end;
  ea = get_inf_attr(INF_MIN_EA);
  end = get_inf_attr(INF_MAX_EA);
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(ea, SEGATTR_BITNESS, 0);
  del_items(ea, DELIT_SIMPLE, end - ea);
  while (ea < end)
  {
    create_insn(ea);
    ea = next_head(ea, end);
    if (ea == BADADDR)
      break;
  }
  out = fopen("/work/pc98-attack-effect-phase-overlay13.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s min=%08X max=%08X\n", get_input_file_path(),
          get_inf_attr(INF_MIN_EA), get_inf_attr(INF_MAX_EA));
  emit_range(out, "POST_HIT_EFFECT_DISPATCH", 0x15E0, 0x1950);
  emit_range(out, "NEWATTACKER", 0x19CB, 0x1A59);
  emit_range(out, "ATTACKE", 0x1A59, 0x1D1C);
  fclose(out);
  qexit(0);
}
