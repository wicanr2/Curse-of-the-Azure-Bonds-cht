#include <idc.idc>

/*
 * Non-destructive PC-98 overlay-14 action-helper audit.
 *
 * The MOVEPARTY caller is already bounded by the raw overlay-14 audit.  This
 * companion report keeps the original overlay-local offsets and bytes while
 * dumping the three near-call helper ranges and the common result helper.
 * It runs only on a disposable IDA database created from the raw overlay; it
 * does not rename, comment, or patch an original executable or baseline .i64.
 * Range labels are navigation aids, not semantic claims about B/P/K.
 */

#define EXPECTED_SHA256 "a8e03ba9a5381c3a9f7ab411ced3262b21e0b65b948160d614386d677610e7b9"

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
  fprintf(out, "range=%s start=0x%04X end=0x%04X\n", label, start, end);
  ea = start;
  while (ea < end)
  {
    size = get_item_size(ea);
    if (size <= 0)
      size = 1;
    fprintf(out, "range=%s local=0x%04X bytes=", label, ea);
    for (index = 0; index < size && ea + index < end; index = index + 1)
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
  out = fopen("/tmp/pc98-overlay14-action-helpers-audit.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", get_input_file_path());
  fprintf(out, "basis=overlay-local code offset; PC-98 Borland MOVEPARTY=00C9:0BCCh\n");
  fprintf(out, "sha256=%s\n", EXPECTED_SHA256);
  fprintf(out, "semantic_status=exact continuous bytes; helper/map meaning remains unknown\n");
  decode_all();
  emit_range(out, "B_HELPER_CANDIDATE", 0x02B0, 0x0558);
  emit_range(out, "P_HELPER_CANDIDATE", 0x0558, 0x06C5);
  emit_range(out, "K_HELPER_CANDIDATE", 0x06D0, 0x0770);
  emit_range(out, "COMMON_RESULT_CANDIDATE", 0x0780, 0x0840);
  emit_range(out, "MOVEPARTY_CALLER", 0x0BCC, 0x0DAA);
  fclose(out);
  qexit(0);
}
