#include <idc.idc>

/* Read-only segment inventory for resident Borland segment:offset resolution. */

static main()
{
  auto input;
  auto seg;
  auto out;

  auto_wait();
  input = get_input_file_path();
  if (strstr(input, "START.EXE") == -1)
    qexit(2);
  out = fopen("/tmp/dos-start-segments.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", input);
  fprintf(out, "basis=IDA segment start/end/name/orgbase/selector; no semantic join\n");
  for (seg = get_first_seg(); seg != BADADDR; seg = get_next_seg(seg))
  {
    fprintf(out, "segment name=%s start=0x%08X end=0x%08X orgbase=0x%08X sel=0x%04X\n",
            get_segm_name(seg), get_segm_start(seg), get_segm_end(seg),
            get_segm_attr(seg, SEGATTR_ORGBASE), get_segm_attr(seg, SEGATTR_SEL));
  }
  fclose(out);
  qexit(0);
}
