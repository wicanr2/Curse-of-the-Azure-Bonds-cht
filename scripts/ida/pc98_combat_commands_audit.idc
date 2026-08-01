#include <idc.idc>

/* Non-destructive PC-98 combat command audit. Run on extracted overlays. */

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
  auto base, end;
  auto_wait();
  set_processor_type("8086", SETPROC_LOADER);
  base = get_inf_attr(INF_MIN_EA);
  end = get_inf_attr(INF_MAX_EA);
  set_segm_attr(base, SEGATTR_BITNESS, 0);
  del_items(base, DELIT_SIMPLE, end - base);
  if (strstr(get_input_file_path(), "overlay-08.bin") != -1)
  {
    auto out8;
    out8 = fopen("/work/pc98-combat-commands-overlay08.txt", "w");
    if (out8 == 0)
      qexit(2);
    emit_range(out8, "SPEED_MENU", 0x1200, 0x1375);
    emit_range(out8, "SET_QUICK_FIGHT", 0x1375, 0x1410);
    fclose(out8);
  }
  else if (strstr(get_input_file_path(), "overlay-13.bin") != -1)
  {
    auto out13;
    out13 = fopen("/work/pc98-combat-commands-overlay13.txt", "w");
    if (out13 == 0)
      qexit(2);
    emit_range(out13, "MOVE_STEP_INTO_GUARD_ATTACK", 0x0684, 0x0790);
    fclose(out13);
  }
  else if (strstr(get_input_file_path(), "overlay-18.bin") != -1)
  {
    auto out18;
    out18 = fopen("/work/pc98-combat-commands-overlay18.txt", "w");
    if (out18 == 0)
      qexit(2);
    emit_range(out18, "ANIMATION_SPEED_CONSUMER", 0x0980, 0x0B88);
    fclose(out18);
  }
  else if (strstr(get_input_file_path(), "overlay-24.bin") != -1)
  {
    auto out24;
    out24 = fopen("/work/pc98-combat-commands-overlay24.txt", "w");
    if (out24 == 0)
      qexit(2);
    emit_range(out24, "CLEAR_ACTIONS", 0x2A5B, 0x2AAE);
    emit_range(out24, "GUARD", 0x2AAE, 0x2B30);
    emit_range(out24, "BANDAGE", 0x35D8, 0x3687);
    fclose(out24);
  }
  else
    qexit(1);
  qexit(0);
}
