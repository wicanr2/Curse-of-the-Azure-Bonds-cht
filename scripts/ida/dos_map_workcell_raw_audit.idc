#include <idc.idc>

/*
 * Non-destructive first pass for a DOS GAME.OVR overlay copy.
 *
 * The input is an overlay-local flat binary.  This pass deliberately does
 * not force instruction boundaries: it records raw little-endian candidates
 * so a later targeted IDA analysis can decide whether each hit is an
 * instruction operand, data, or coincidence.
 */

static is_target(value)
{
  value = value & 0xFFFF;
  return value == 0x4BF0 || value == 0x4BF1 ||
         value == 0x4C01 || value == 0x4C28 || value == 0x4C29 ||
         value == 0x4C2A || value == 0x4C2B || value == 0x4C2C ||
         value == 0x4CBA || value == 0x4CBB || value == 0x4CC0 ||
         value == 0x7F70 || value == 0x7F71 || value == 0x7F72 ||
         value == 0x7F82;
}

static emit_context(out, base, ea)
{
  auto start;
  auto end;
  auto p;

  start = ea - 12;
  if (start < base)
    start = base;
  end = ea + 14;
  if (end > get_inf_attr(INF_MAX_EA))
    end = get_inf_attr(INF_MAX_EA);
  fprintf(out, "context_local=0x%04X bytes=", start - base);
  for (p = start; p < end; p = p + 1)
    fprintf(out, "%02X", get_wide_byte(p));
  fprintf(out, "\n");
}

static main()
{
  auto segment;
  auto base;
  auto end;
  auto ea;
  auto word;
  auto hits;
  auto out;

  segment = get_first_seg();
  if (segment == BADADDR)
    qexit(2);
  set_segm_addressing(segment, 0);
  base = get_segm_start(segment);
  end = get_segm_end(segment);
  out = fopen("/tmp/dos-map-workcell-raw-audit.txt", "w");
  if (out == 0)
    qexit(2);
  fprintf(out, "input=%s\n", get_input_file_path());
  fprintf(out, "basis=overlay-local raw file offset; no forced instruction boundaries\n");
  fprintf(out, "targets=4BF0,4BF1,4C01,4C28..4C2C,4CBA,4CBB,4CC0,7F70..7F72,7F82\n");
  hits = 0;
  ea = base;
  while (ea + 1 < end)
  {
    word = get_wide_word(ea);
    if (is_target(word))
    {
      fprintf(out, "RAW_LE16 local=0x%04X target=0x%04X bytes=%02X%02X\n",
              ea - base, word & 0xFFFF, get_wide_byte(ea),
              get_wide_byte(ea + 1));
      emit_context(out, base, ea);
      hits = hits + 1;
    }
    ea = ea + 1;
  }
  fprintf(out, "RAW_LE16_COUNT=%d\n", hits);
  fclose(out);
  qexit(0);
}
