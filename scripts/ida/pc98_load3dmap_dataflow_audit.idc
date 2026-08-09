#include <idc.idc>

/*
 * Non-destructive PC-98 LOAD3DMAP boundary audit.
 *
 * The appended Borland symbol table identifies LOAD3DMAP as 017C:1253h.
 * This script only decodes a disposable raw overlay copy and emits the
 * continuous range from that symbol boundary.  It does not rename or patch
 * a baseline database, and it does not infer a secret-door writer from a
 * loader call or a matching numeric constant.
 */

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
    for (index = 0; index < size && ea + index < end; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(ea + index));
    fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
    ea = ea + size;
  }
}

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

static main()
{
  auto out;

  auto_wait();
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(get_inf_attr(INF_MIN_EA), SEGATTR_BITNESS, 0);
  out = fopen("/tmp/pc98-load3dmap-dataflow-audit.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s min=0x%08X max=0x%08X\n", get_input_file_path(),
          get_inf_attr(INF_MIN_EA), get_inf_attr(INF_MAX_EA));
  fprintf(out, "basis=overlay-local code offset; Borland LOAD3DMAP=017C:1253h\n");
  fprintf(out, "semantic_status=exact continuous bytes; map/secret-door meaning remains unknown\n");
  decode_all();
  fprintf(out, "-- BLOCKCODE 017C:04DEh..060Dh --\n");
  emit_range(out, 0x04DE, 0x060D);
  fprintf(out, "-- WALLCODE 017C:060Dh..0710h --\n");
  emit_range(out, 0x060D, 0x0710);
  fprintf(out, "-- LOAD3DMAP 017C:1253h..end --\n");
  emit_range(out, 0x1253, get_inf_attr(INF_MAX_EA));
  fclose(out);
  qexit(0);
}
