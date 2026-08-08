#include <idc.idc>

/*
 * Non-destructive PC-98 Quick target caller audit.
 *
 * Run only against a copied code-only overlay-09 image. The report preserves
 * overlay-local addresses and raw call-site bytes; it never renames or patches
 * the pristine GAME.OVR image or a baseline IDA database.
 */

static signed_word(value)
{
  if ((value & 0x8000) != 0)
    return value - 0x10000;
  return value;
}

static decode_all()
{
  auto base;
  auto end;
  auto ea;
  auto size;

  base = get_inf_attr(INF_MIN_EA);
  end = get_inf_attr(INF_MAX_EA);
  del_items(base, DELIT_SIMPLE, end - base);
  ea = base;
  while (ea < end)
  {
    size = create_insn(ea);
    if (size <= 0)
      size = 1;
    ea = ea + size;
  }
}

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
    fprintf(out, "label=%s ea=%08X local=%04X bytes=", label, ea,
            ea - base);
    for (index = 0; index < size; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(ea + index));
    fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
    ea = ea + size;
  }
}

static emit_xrefs(out, base, local, label)
{
  auto target;
  auto x;
  auto count;

  target = base + local;
  count = 0;
  for (x = get_first_cref_to(target); x != BADADDR;
       x = get_next_cref_to(target, x))
  {
    fprintf(out, "%s_XREF ea=%08X local=%04X type=%d bytes=", label,
            x, x - base, XrefType());
    fprintf(out, "%02X%02X%02X", get_wide_byte(x), get_wide_byte(x + 1),
            get_wide_byte(x + 2));
    fprintf(out, " disasm=%s\n", generate_disasm_line(x, 0));
    count = count + 1;
  }
  fprintf(out, "%s_XREF count=%d\n", label, count);
}

static scan_raw_calls(out, base)
{
  auto end;
  auto ea;
  auto local_target;
  auto offset;

  end = get_inf_attr(INF_MAX_EA);
  ea = base;
  while (ea + 2 < end)
  {
    if (get_wide_byte(ea) == 0xE8)
    {
      offset = signed_word(get_wide_word(ea + 1));
      local_target = (ea + 3 + offset) - base;
      if (local_target == 0x04CC || local_target == 0x03D3 ||
          local_target == 0x0627)
      {
        fprintf(out, "RAW_NEAR_CALL ea=%08X local=%04X target_local=%04X bytes=",
                ea, ea - base, local_target);
        fprintf(out, "%02X%02X%02X", get_wide_byte(ea), get_wide_byte(ea + 1),
                get_wide_byte(ea + 2));
        fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
      }
    }
    if (get_wide_byte(ea) == 0x9A && get_wide_word(ea + 1) == 0x04CC)
    {
      fprintf(out, "RAW_FAR_OFFSET_CALL ea=%08X local=%04X bytes=",
              ea, ea - base);
      fprintf(out, "%02X%02X%02X%02X%02X", get_wide_byte(ea),
              get_wide_byte(ea + 1), get_wide_byte(ea + 2),
              get_wide_byte(ea + 3), get_wide_byte(ea + 4));
      fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
    }
    ea = ea + 1;
  }
}

static main()
{
  auto path;
  auto out;
  auto base;

  auto_wait();
  path = get_input_file_path();
  if (strstr(path, "overlay-09.bin") == -1)
    qexit(1);
  base = get_inf_attr(INF_MIN_EA);
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(base, SEGATTR_BITNESS, 0);
  out = fopen("/work/pc98-quick-target-xrefs.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s min=%08X max=%08X\n", path, base,
          get_inf_attr(INF_MAX_EA));
  decode_all();
  emit_range(out, "TARGET_HELPER_CALLER_CONTEXT", 0x0000, 0x02D3);
  emit_range(out, "QUICK_TARGET_HELPER_CONTEXT", 0x03D3, 0x0755);
  emit_range(out, "LOCAL_HANDOFF_TARGETS", 0x1380, 0x15C0);
  emit_xrefs(out, base, 0x03D3, "SUITABILITY");
  emit_xrefs(out, base, 0x04CC, "TARGET_HELPER");
  emit_xrefs(out, base, 0x0627, "QUICK_AI");
  scan_raw_calls(out, base);
  fclose(out);
  qexit(0);
}
