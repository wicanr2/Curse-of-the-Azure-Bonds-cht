#include <idc.idc>

/*
 * Non-destructive IDA Pro 9.4 audit for PC-98 combat spell targeting.
 * Inputs are read-only code-only overlay copies. Reports retain original
 * overlay-local addresses and bytes; no semantic rename is written back.
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

  if (strstr(path, "overlay-08.bin") != -1)
    output_path = "/work/pc98-spell-targeting-overlay08.txt";
  else if (strstr(path, "overlay-13.bin") != -1)
    output_path = "/work/pc98-spell-targeting-overlay13.txt";
  else if (strstr(path, "overlay-22.bin") != -1)
    output_path = "/work/pc98-spell-targeting-overlay22.txt";
  else if (strstr(path, "overlay-24.bin") != -1)
    output_path = "/work/pc98-spell-targeting-overlay24.txt";
  else if (strstr(path, "overlay-32.bin") != -1)
    output_path = "/work/pc98-spell-targeting-overlay32.txt";
  else if (strstr(path, "overlay-31.bin") != -1)
    output_path = "/work/pc98-spell-targeting-overlay31.txt";
  else
    qexit(1);

  out = fopen(output_path, "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s min=%08X max=%08X\n", path, base, end);

  if (strstr(path, "overlay-08.bin") != -1)
    emit_range(out, "COMBAT_TARGET_DELEGATE_WRITER", 0x00A0, 0x0190);
  else if (strstr(path, "overlay-13.bin") != -1)
  {
    emit_range(out, "CHECKTARGET", 0x1160, 0x1280);
    emit_range(out, "COMBAT_SPELL_TARGETING", 0x225F, 0x27A1);
    emit_range(out, "CASTCOMBATSPELL", 0x27A1, 0x2915);
    emit_range(out, "PICKTARGET", 0x3D20, 0x3FA0);
  }
  else if (strstr(path, "overlay-22.bin") != -1)
  {
    emit_range(out, "FIGSPELLRANGE", 0x0D80, 0x0E90);
    emit_range(out, "GETSPELLTARGETS", 0x10E0, 0x1250);
    emit_range(out, "TARGETDIR", 0x1240, 0x1340);
  }
  else if (strstr(path, "overlay-24.bin") != -1)
    emit_range(out, "BUILD_NEAR_TARGETS_CANDIDATE", 0x2820, 0x29C0);
  else if (strstr(path, "overlay-31.bin") != -1)
  {
    emit_range(out, "LOS_SCAN_STABLE_SORT", 0x0035, 0x019D);
    emit_range(out, "LOS_SCAN_TARGET_LIST", 0x08D8, 0x0BA5);
  }
  else
    emit_range(out, "REBUILD_SORTED_COMBATANT_LIST_CANDIDATE", 0x00F0, 0x0430);
  fclose(out);
  qexit(0);
}
