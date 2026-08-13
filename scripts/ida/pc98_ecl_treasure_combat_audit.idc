#include <idc.idc>

/*
 * Non-destructive PC-98 ECL TREASURE/COMBAT boundary audit.
 *
 * Inputs are disposable raw copies of Borland INTERPET overlay 2 or POSTCOM
 * overlay 5.  The script preserves overlay-local offsets and raw bytes; labels
 * below describe only the range being inspected and do not rename functions or
 * assert semantics.
 */

static define_function(start)
{
  del_items(start, DELIT_SIMPLE, 1);
  create_insn(start);
  add_func(start, BADADDR);
  auto_wait();
}

static define_bounded_function(start, end)
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
  add_func(start, end);
  auto_wait();
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

static emit_function(out, label, start)
{
  auto ea;
  auto end;
  auto size;
  auto index;

  end = get_func_attr(start, FUNCATTR_END);
  fprintf(out, "function=%s start=0x%04X end=0x%04X\n", label, start, end);
  if (end == BADADDR)
    return;
  ea = start;
  while (ea < end)
  {
    size = get_item_size(ea);
    if (size <= 0)
      size = 1;
    fprintf(out, "function=%s local=0x%04X bytes=", label, ea);
    for (index = 0; index < size; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(ea + index));
    fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
    ea = ea + size;
  }
}

static main()
{
  auto out;
  auto end;

  auto_wait();
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(get_inf_attr(INF_MIN_EA), SEGATTR_BITNESS, 0);
  out = fopen("/tmp/pc98-ecl-boundary-audit.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", get_input_file_path());
  fprintf(out, "basis=overlay-local code offset; raw disposable database\n");
  fprintf(out, "semantic_status=raw bytes only; interpretation requires external ledger\n");
  end = get_inf_attr(INF_MAX_EA);
  fprintf(out, "min=0x%04X max=0x%04X\n", get_inf_attr(INF_MIN_EA), end);
  if (end > 0x3000)
  {
    define_function(0x1820);
    define_bounded_function(0x1BEA, 0x2125);
    define_bounded_function(0x3A21, 0x3CEB);
    emit_function(out, "INTERPET_OPCODE_24_HANDLER", 0x1820);
    emit_function(out, "INTERPET_OPCODE_27_HANDLER", 0x1BEA);
    emit_function(out, "INTERPET_GOECL", 0x3A21);
    emit_range(out, "INTERPET_OPCODE_DISPATCH", 0x37F0, 0x39B0);
  }
  else
  {
    define_bounded_function(0x1775, 0x19A3);
    emit_function(out, "POSTCOM_DOPOSTCOMBAT", 0x1775);
  }
  fclose(out);
  qexit(0);
}
