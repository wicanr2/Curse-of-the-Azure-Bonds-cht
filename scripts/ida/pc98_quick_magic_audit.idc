#include <idc.idc>

/*
 * Non-destructive PC-98 combat QUICK/MAGIC audit.
 *
 * Run only against an extracted copy of an accepted overlay binary.  The report preserves
 * raw local offsets and bytes; labels below are research scopes, not names
 * written back to the pristine GAME.OVR image.
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
  auto base, end, out;
  auto_wait();
  set_processor_type("8086", SETPROC_LOADER);
  base = get_inf_attr(INF_MIN_EA);
  end = get_inf_attr(INF_MAX_EA);
  set_segm_attr(base, SEGATTR_BITNESS, 0);
  del_items(base, DELIT_SIMPLE, end - base);
  if (strstr(get_input_file_path(), "overlay-08.bin") != -1)
  {
    out = fopen("/work/pc98-quick-magic-overlay08.txt", "w");
    if (out == 0)
      qexit(2);
    emit_range(out, "COMBAT_INPUT_AND_QUICK_HANDOFF", 0x0200, 0x0780);
    emit_range(out, "PENDING_SPELL_ACTION_CONSUMER", 0x0414, 0x0475);
    emit_range(out, "SET_QUICK_FIGHT", 0x1375, 0x13C6);
  }
  else if (strstr(get_input_file_path(), "overlay-09.bin") != -1)
  {
    out = fopen("/work/pc98-quick-magic-overlay09.txt", "w");
    if (out == 0)
      qexit(2);
    emit_range(out, "QUICK_SPELL_MIN_RANGE_PREDICATE", 0x02D3, 0x03D3);
    emit_range(out, "QUICK_SPELL_SUITABILITY_HYPOTHESIS", 0x03B0, 0x04CC);
    emit_range(out, "QUICK_TARGET_SELECTION_HYPOTHESIS", 0x04CC, 0x0627);
    emit_range(out, "MAGIC_FLAG_QUICK_AI_CONSUMER", 0x0627, 0x07C0);
    emit_range(out, "MAGIC_FLAG_AI_CONSUMER", 0x1290, 0x1340);
  }
  else if (strstr(get_input_file_path(), "overlay-10.bin") != -1)
  {
    out = fopen("/work/pc98-quick-magic-overlay10.txt", "w");
    if (out == 0)
      qexit(2);
    emit_range(out, "MAGIC_FLAG_SECONDARY_CONSUMER", 0x1C40, 0x1CD0);
  }
  else if (strstr(get_input_file_path(), "overlay-13.bin") != -1)
  {
    out = fopen("/work/pc98-quick-magic-overlay13.txt", "w");
    if (out == 0)
      qexit(2);
    emit_range(out, "STUB_00B8_0075_HANDLER_HYPOTHESIS", 0x1E30, 0x2050);
  }
  else if (strstr(get_input_file_path(), "overlay-17.bin") != -1)
  {
    out = fopen("/work/pc98-quick-magic-overlay17.txt", "w");
    if (out == 0)
      qexit(2);
    emit_range(out, "BORLAND_CANDOCURE", 0x36BD, 0x37AB);
    emit_range(out, "BORLAND_DOCURE", 0x39CA, 0x3B40);
  }
  else
    qexit(1);
  fclose(out);
  qexit(0);
}
