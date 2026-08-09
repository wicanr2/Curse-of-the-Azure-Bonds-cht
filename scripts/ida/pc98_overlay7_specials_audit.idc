#include <idc.idc>

/*
 * Non-destructive PC-98 ECL2 overlay audit for CHECKSPECIALS/STORESPECIALS.
 * These are the named map-special entry points in Borland segment 0062h;
 * the dump keeps raw overlay-local offsets and does not turn names into
 * semantic proof of a secret-door implementation.
 */

static decode_all()
{
  auto ea;
  auto end;
  auto size;

  del_items(0, DELIT_SIMPLE, get_inf_attr(INF_MAX_EA));
  ea = 0;
  end = get_inf_attr(INF_MAX_EA);
  while (ea < end)
  {
    size = create_insn(ea);
    if (size <= 0)
      size = 1;
    ea = ea + size;
  }
}

static emit_range(out, label, start, end)
{
  auto ea;
  auto size;
  auto index;

  if (end > get_inf_attr(INF_MAX_EA))
    end = get_inf_attr(INF_MAX_EA);
  ea = start;
  while (ea < end)
  {
    size = get_item_size(ea);
    if (size <= 0)
      size = 1;
    fprintf(out, "range=%s local=0x%04X bytes=", label, ea);
    for (index = 0; index < size; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(ea + index));
    fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
    ea = ea + size;
  }
}

static main()
{
  auto out;

  auto_wait();
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(get_inf_attr(INF_MIN_EA), SEGATTR_BITNESS, 0);
  out = fopen("/tmp/pc98-overlay7-specials-audit.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", get_input_file_path());
  fprintf(out, "basis=overlay-local code offset; Borland ECL2 segment=0062h\n");
  fprintf(out, "named_targets=CHECKSPECIALS:08BDh STORESPECIALS:0C28h\n");
  fprintf(out, "semantic_status=exact bytes／control flow; search bridge requires caller／runtime closure\n");
  decode_all();
  emit_range(out, "CHECKSPECIALS", 0x0800, 0x0C80);
  emit_range(out, "STORESPECIALS", 0x0C00, 0x1400);
  fclose(out);
  qexit(0);
}
