#include <idc.idc>

/*
 * Non-destructive PC-98 overlay 2 (Borland INTERPET segment 0037) audit.
 *
 * This is a candidate-state report for the shared BDF0/BDF1 and character
 * +594h fields used by movement.  Raw local offsets and bytes remain the
 * evidence; names below are module/address labels only and do not assert
 * semantics without the MOVEMENT caller and runtime bridge.
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
  out = fopen("/tmp/pc98-overlay2-search-state-audit.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", get_input_file_path());
  fprintf(out, "basis=overlay-local code offset; Borland module=INTERPET/segment0037h\n");
  fprintf(out, "semantic_status=unknown／candidate state flow only\n");
  decode_all();
  emit_range(out, "BDF0_INIT_AND_FIRST_USES", 0x0B80, 0x0D20);
  emit_range(out, "OBJECT_594_CONTEXT", 0x1A80, 0x1B80);
  emit_range(out, "BDF0_LATE_USES", 0x3200, 0x3C20);
  fclose(out);
  qexit(0);
}
