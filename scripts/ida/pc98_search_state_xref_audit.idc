#include <idc.idc>

/*
 * Non-destructive PC-98 SEARCH/TEMPSEARCH xref audit.
 *
 * The appended Borland symbols identify TEMPSEARCH as 0C29:BDF1.  In the
 * baseline PC-98 database that Borland data segment is mapped at IDA linear
 * 01C290h, so BDF0/BDF1 resolve to 028080h/028081h.  Keep both address spaces
 * in the report.  This script only reads the existing database: it does not
 * rename, comment, decode, or delete any item.  It therefore cannot turn a
 * direct dref into a secret-door semantic claim by itself.
 */

static emit_window(out, label, center, before, after)
{
  auto start;
  auto end;
  auto ea;
  auto size;
  auto index;

  start = center - before;
  end = center + after;
  if (start < get_inf_attr(INF_MIN_EA))
    start = get_inf_attr(INF_MIN_EA);
  if (end > get_inf_attr(INF_MAX_EA))
    end = get_inf_attr(INF_MAX_EA);
  fprintf(out, "window=%s start=0x%08X end=0x%08X\n", label, start, end);
  ea = start;
  while (ea < end)
  {
    size = get_item_size(ea);
    if (size <= 0)
      size = 1;
    fprintf(out, "window_item=%s ea=0x%08X bytes=", label, ea);
    for (index = 0; index < size && ea + index < end; index = index + 1)
      fprintf(out, "%02X", get_wide_byte(ea + index));
    fprintf(out, " disasm=%s\n", generate_disasm_line(ea, 0));
    ea = ea + size;
  }
}

static emit_function(out, from)
{
  auto start;
  auto end;

  start = get_func_attr(from, FUNCATTR_START);
  end = get_func_attr(from, FUNCATTR_END);
  if (start == BADADDR || end == BADADDR || end <= start)
  {
    fprintf(out, "function from=0x%08X status=unavailable\n", from);
    return;
  }
  fprintf(out, "function from=0x%08X start=0x%08X end=0x%08X name=%s\n",
          from, start, end, get_func_name(from));
}

static emit_target(out, label, borland_offset, ida_ea)
{
  auto x;
  auto count;
  auto type;

  fprintf(out, "target label=%s borland=0C29:%04X ida_ea=0x%08X ida_name=%s bytes=",
          label, borland_offset, ida_ea, get_name(ida_ea));
  fprintf(out, "%02X%02X%02X%02X\n", get_wide_byte(ida_ea),
          get_wide_byte(ida_ea + 1), get_wide_byte(ida_ea + 2),
          get_wide_byte(ida_ea + 3));

  count = 0;
  for (x = get_first_dref_to(ida_ea); x != BADADDR;
       x = get_next_dref_to(ida_ea, x))
  {
    type = XrefType();
    fprintf(out, "dref label=%s target=0x%08X from=0x%08X type=%d name=%s disasm=%s\n",
            label, ida_ea, x, type, get_name(x), generate_disasm_line(x, 0));
    emit_function(out, x);
    emit_window(out, label, x, 0x20, 0x40);
    count = count + 1;
  }
  fprintf(out, "dref label=%s count=%d\n", label, count);

  count = 0;
  for (x = get_first_cref_to(ida_ea); x != BADADDR;
       x = get_next_cref_to(ida_ea, x))
  {
    type = XrefType();
    fprintf(out, "cref label=%s target=0x%08X from=0x%08X type=%d name=%s disasm=%s\n",
            label, ida_ea, x, type, get_name(x), generate_disasm_line(x, 0));
    emit_function(out, x);
    emit_window(out, label, x, 0x20, 0x40);
    count = count + 1;
  }
  fprintf(out, "cref label=%s count=%d\n", label, count);
}

static main()
{
  auto out;

  auto_wait();
  out = fopen("/tmp/pc98-search-state-xref-audit.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s min=0x%08X max=0x%08X\n", get_input_file_path(),
          get_inf_attr(INF_MIN_EA), get_inf_attr(INF_MAX_EA));
  fprintf(out, "address_basis=borland_segment_0C29 ida_linear_base=0x0001C290\n");
  fprintf(out, "semantic_status=exact direct xref inventory; secret-door meaning remains unknown\n");
  emit_target(out, "CURRENTECL_BDF0", 0xBDF0, 0x28080);
  emit_target(out, "TEMPSEARCH_BDF1", 0xBDF1, 0x28081);
  emit_window(out, "SEARCH_STATE_NEIGHBORHOOD", 0x28080, 0x20, 0x40);
  fclose(out);
  qexit(0);
}
