#include <idc.idc>

/*
 * Non-destructive audit of PC-98 overlay 24 (Borland segment 014A).
 *
 * PREMOVEPARTY calls the resident stub 014A:00DE.  Typed TPOV resolution
 * maps that stub to overlay-24 entry 38 and handler-local 2E8Ch; the numeric
 * stub offset is not the code offset.  The segment and overlay-local basis
 * are kept explicit; this script does not rename or patch the pristine
 * PC98-GAME.EXE/OVR or any baseline database.
 */

static decode_range(start, end)
{
  auto ea;
  auto size;

  del_items(start, DELIT_SIMPLE, end - start);
  ea = start;
  while (ea < end)
  {
    size = create_insn(ea);
    if (size <= 0)
      size = 1;
    ea = ea + size;
  }
}

static emit_range(out, start, end)
{
  auto ea;
  auto size;
  auto index;

  ea = start;
  while (ea < end)
  {
    size = get_item_size(ea);
    if (size <= 0)
      size = 1;
    fprintf(out, "local=0x%04X bytes=", ea);
    for (index = 0; index < size; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(ea + index));
    fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
    ea = ea + size;
  }
}

static emit_raw_far_calls(out, target_offset, target_segment)
{
  auto ea;
  auto end;

  end = get_inf_attr(INF_MAX_EA);
  for (ea = get_inf_attr(INF_MIN_EA); ea + 4 < end; ea = ea + 1)
  {
    if (get_wide_byte(ea) == 0x9A &&
        get_wide_word(ea + 1) == target_offset &&
        get_wide_word(ea + 3) == target_segment)
    {
      fprintf(out, "RAW_FAR_CALL local=0x%04X target=014A:00DE bytes=%02X%02X%02X%02X%02X\n",
              ea, get_wide_byte(ea), get_wide_byte(ea + 1),
              get_wide_byte(ea + 2), get_wide_byte(ea + 3),
              get_wide_byte(ea + 4));
    }
  }
}

static main()
{
  auto out;
  auto end;

  auto_wait();
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(get_inf_attr(INF_MIN_EA), SEGATTR_BITNESS, 0);
  out = fopen("/tmp/pc98-overlay24-generic-audit.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", get_input_file_path());
  fprintf(out, "basis=overlay-local code offset; Borland stub=014A:00DEh; TPOV entry=38; handler=0x2E8C\n");
  fprintf(out, "semantic_status=unknown until S caller／map writer／runtime bridge is closed\n");
  end = get_inf_attr(INF_MAX_EA);
  decode_range(0x0000, end);
  emit_range(out, 0x2C80, 0x3100);
  emit_raw_far_calls(out, 0x00DE, 0x014A);
  fclose(out);
  qexit(0);
}
