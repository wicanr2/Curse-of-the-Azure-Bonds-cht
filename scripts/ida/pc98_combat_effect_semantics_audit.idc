#include <idc.idc>

/*
 * Non-destructive PC-98 combat/effect semantic windows.
 *
 * Run on code-only overlay copies extracted from the verified GAME.OVR. The
 * labels are research scopes; original symbols, bytes and overlay-local
 * addresses remain unchanged. Any semantic ledger must cite the source
 * executable hash and keep a confidence level beside these ranges.
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
  base = get_inf_attr(INF_MIN_EA);
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(base, SEGATTR_BITNESS, 0);

  if (strstr(path, "overlay-13.bin") != -1)
  {
    out = fopen("/work/pc98-combat-effect-overlay13.txt", "w");
    if (out == 0)
      qexit(2);
    fprintf(out, "input=%s min=%08X max=%08X\n", path, base,
            get_inf_attr(INF_MAX_EA));
    emit_range(out, "FIGGOODBAD", 0x2700, 0x2A00);
    emit_range(out, "FIGRANGE", 0x2970, 0x2A50);
    emit_range(out, "FIGCASTERLEVEL", 0x2A80, 0x2B40);
    emit_range(out, "EFFECT_WRITER_STUB_TARGET", 0x2D70, 0x2E20);
    emit_range(out, "COMBAT_EFFECT_CALLERS", 0x1500, 0x1A80);
  }
  else if (strstr(path, "overlay-12.bin") != -1)
  {
    out = fopen("/work/pc98-combat-effect-overlay12.txt", "w");
    if (out == 0)
      qexit(2);
    fprintf(out, "input=%s min=%08X max=%08X\n", path, base,
            get_inf_attr(INF_MAX_EA));
    emit_range(out, "PROTECTED_AND_DAMAGE_HANDLERS", 0x001B, 0x00C0);
    emit_range(out, "DAMAGE_EFFECT_WINDOW", 0x2300, 0x2540);
  }
  else
    qexit(1);
  fclose(out);
  qexit(0);
}
