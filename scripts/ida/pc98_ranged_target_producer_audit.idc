#include <idc.idc>

/*
 * Non-destructive IDA Pro 9.4 audit for the PC-98 ranged-target producer.
 *
 * Run only against a copied code-only overlay-24 image. The report keeps the
 * original overlay-local address and raw instruction bytes. It deliberately
 * does not rename functions, infer field names, or modify a baseline .i64.
 */

static emit_range(out, label, relative_start, relative_end)
{
  auto base, start, end, ea, size, index;
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

static main()
{
  auto path, output_path, out, base, end, ea, size;
  auto_wait();
  path = get_input_file_path();
  if (strstr(path, "overlay-24.bin") == -1)
    qexit(1);
  output_path = "/work/pc98-ranged-target-producer.txt";
  out = fopen(output_path, "w");
  if (out == 0)
    qexit(2);

  base = get_inf_attr(INF_MIN_EA);
  end = get_inf_attr(INF_MAX_EA);
  set_processor_type("8086", SETPROC_LOADER);
  set_segm_attr(base, SEGATTR_BITNESS, 0);
  del_items(base, DELIT_SIMPLE, end - base);
  ea = base;
  while (ea < end)
  {
    size = create_insn(ea);
    if (size <= 0)
      size = 1;
    ea = ea + size;
  }

  fprintf(out, "input=%s min=%08X max=%08X\n", path, base, end);
  /* Existing report stopped at 29C0h; keep the continuation through 2C80h. */
  emit_range(out, "TARGET_PRODUCER_INIT_AND_FIELDS", 0x2820, 0x2C80);
  fclose(out);
  qexit(0);
}
