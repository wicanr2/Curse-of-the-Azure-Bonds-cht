#!/usr/bin/env python3
"""從原版擷取畫面讀出「x y 朝向」（讀不到就印三個問號）。"""
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dos_screen import read_screen  # noqa: E402

match = re.search(r"(\d+),(\d+) ([NESW])", " ".join(read_screen(sys.argv[1])))
print("\t".join(match.groups()) if match else "?\t?\t?")
