#include <idc.idc>

/*
 * Non-destructive PC-98 overlay 11 audit for the BDF0/BDF1 bridge.
 *
 * This report is intentionally narrow: it preserves overlay-local offsets and
 * raw bytes around the symbol-independent shared state.  A BDF1 read/write or
 * a +594h copy is not by itself proof of a secret-door map mutation.
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

static emit_range(out, start, end)
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
    fprintf(out, "local=0x%04X bytes=", ea);
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
  out = fopen("/tmp/pc98-overlay11-bdf1-audit.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", get_input_file_path());
  fprintf(out, "basis=overlay-local code offset; Borland module=INIT/segment004Ah\n");
  fprintf(out, "semantic_status=unknown until BDF1 writer-to-map consumer is closed\n");
  decode_all();
  emit_range(out, 0x0280, 0x0500);
  emit_range(out, 0x05C0, 0x0760);
  fclose(out);
  qexit(0);
}
