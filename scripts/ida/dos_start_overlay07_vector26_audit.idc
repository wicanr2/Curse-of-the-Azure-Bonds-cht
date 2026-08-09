#include <idc.idc>

/*
 * Read-only audit of the START.EXE control block that dispatches
 * overlay-07 vector 26 to local 1B3F.  The control block is addressed in
 * the IDA image, but the report keeps raw file offset, IDA EA, runtime
 * selector, vector index, and overlay-local target separate.
 */

static emit_bytes(out, ea, count)
{
  auto index;

  for (index = 0; index < count; index = index + 1)
    fprintf(out, "%02X", get_wide_byte(ea + index));
}

static emit_vector(out, control_ea, index)
{
  auto ea;

  ea = control_ea + 0x20 + index * 5;
  fprintf(out, "vector index=%d control_file_offset=0x%04X ida_ea=0x%08X bytes=",
          index, 0x0E60 + 0x20 + index * 5, ea);
  emit_bytes(out, ea, 5);
  fprintf(out, " target_local=0x%04X\n", get_wide_word(ea + 2));
}

static main()
{
  auto input;
  auto out;
  auto segment;
  auto control_ea;
  auto index;

  auto_wait();
  input = get_input_file_path();
  if (strstr(input, "START.EXE") == -1)
    qexit(2);
  control_ea = 0x10000 + (0x0E60 - 0x07B0);
  segment = BADADDR;
  for (segment = get_first_seg(); segment != BADADDR;
       segment = get_next_seg(segment))
  {
    if (get_segm_start(segment) <= control_ea && control_ea < get_segm_end(segment))
      break;
  }
  if (segment == BADADDR)
    qexit(2);
  out = fopen("/tmp/dos-start-overlay07-vector26.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=MZ raw control block mapped to IDA image base 0x10000\n");
  fprintf(out, "semantic_status=exact control-vector bytes／target offset; caller meaning remains evidence-scoped\n");
  fprintf(out, "mz_header_size=0x07B0 control_file_offset=0x0E60 control_runtime_selector=0x006B control_ida_ea=0x%08X selector=0x%04X segment=%s\n",
          control_ea, get_segm_attr(segment, SEGATTR_SEL), get_segm_name(segment));
  fprintf(out, "control_header bytes=");
  emit_bytes(out, control_ea, 0x20);
  fprintf(out, "\nvector_table_file_offset=0x0E80 ida_ea=0x%08X entry_size=5\n",
          control_ea + 0x20);
  for (index = 0; index < 33; index = index + 1)
    emit_vector(out, control_ea, index);
  fclose(out);
  qexit(0);
}
