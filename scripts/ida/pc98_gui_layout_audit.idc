#include <idc.idc>

/*
 * IDA Pro 9.4 audit for the PC-98 Gold Box GUI overlay routines.
 *
 * Run against code-only overlays emitted by cmd/pc98-ovr-audit.  The
 * GAME.EXE Borland table maps DRAWWIN, PORTRAIT and THREED to overlay indices
 * 28, 29 and 30.  Commercial code and generated databases remain local.
 */

static make_function(label, relative_start, relative_end)
{
  auto base;
  auto start;
  auto end;
  auto ea;
  auto size;

  base = get_inf_attr(INF_MIN_EA);
  start = base + relative_start;
  end = base + relative_end;
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
  set_name(start, label, SN_FORCE);
  msg("FUNCTION name=%s offset=%04X start=%08X end=%08X\n",
      label, relative_start, start, end);
  for (ea = start; ea < end; ea = next_head(ea, end))
  {
    if (ea == BADADDR)
      break;
    msg("%08X bytes=", ea);
    for (size = 0; size < get_item_size(ea); size = size + 1)
      msg("%02X", get_wide_byte(ea + size));
    msg(" asm=%s\n", generate_disasm_line(ea, 0));
  }
}

static main()
{
  auto path;
  auto maximum;
  auto base;

  auto_wait();
  path = get_input_file_path();
  base = get_inf_attr(INF_MIN_EA);
  maximum = get_inf_attr(INF_MAX_EA);
  /* Raw binary loader defaults can retain a 64-bit segment even with
     -p8086.  Force 16-bit segment decoding before creating any instruction. */
  set_segm_attr(base, SEGATTR_BITNESS, 0);
  del_items(base, DELIT_SIMPLE, maximum - base);
  msg("GUI_LAYOUT input=%s min=%08X max=%08X\n", path, base, maximum);

  if (strstr(path, "overlay-28.bin") != -1)
  {
    make_function("LOADDRAWWIN", 0x0000, 0x0016);
    make_function("DRAWWINDOW", 0x0016, maximum - base);
  }
  else if (strstr(path, "overlay-29.bin") != -1)
  {
    make_function("LOADPORTRAIT", 0x0000, 0x000C);
    make_function("SHOWPORTRAIT", 0x000C, 0x0091);
    make_function("SHOWHEAD", 0x0091, 0x00B3);
    make_function("SHOWBODY", 0x00B3, 0x010B);
    make_function("LOADSEQUENCE", 0x010B, 0x045A);
    make_function("DISPOSESEQUENCE", 0x045A, 0x04E0);
    make_function("LOADCHARACTERPORTRAIT", 0x04E0, 0x05F9);
    make_function("SHOWCHARACTERPORTRAIT", 0x05F9, 0x066C);
    make_function("SHOW3DSPRITE", 0x066C, 0x072B);
    make_function("LOADBIGPIC", 0x072B, 0x0777);
    make_function("SHOWBIGPIC", 0x0777, maximum - base);
  }
  else if (strstr(path, "overlay-30.bin") != -1)
  {
    make_function("LOADTHREED", 0x0000, 0x016B);
    make_function("SET3DCOLORS", 0x016B, 0x018C);
    make_function("CLEAR3DVIEW", 0x018C, 0x04DE);
    make_function("BLOCKCODE", 0x04DE, 0x060D);
    make_function("WALLCODE", 0x060D, 0x0710);
    make_function("SPECIALCODE", 0x0710, 0x078B);
    make_function("BUILDVIEW", 0x078B, 0x0F8F);
    make_function("LOADWALLSET", 0x0F8F, 0x1253);
    make_function("LOAD3DMAP", 0x1253, maximum - base);
  }
  else
  {
    msg("GUI_LAYOUT unsupported_input=1\n");
    qexit(1);
  }
  qexit(0);
}
