#!/bin/sh
# fixture: mirrors the real /sys/class/drm/shape; values synthetic
root=testdata/drm

# card1: the full card (dGPU-like - all sensors)
d=$root/card1/device; h=$d/hwmon/hwmon4
mkdir -p "$h"
printf '42\n'         > "$d/gpu_busy_percent"
printf '1073741824\n' > "$d/mem_info_vram_used"   # 1 GiB
printf '8589934592\n' > "$d/mem_info_vram_total"  # 8 GiB
printf '268425456\n'  > "$d/mem_info_gtt_used"
printf '45000\n'      > "$h/temp1_input"          # 45 °C - proves ÷ 1000
printf '52000\n'      > "$h/temp2_input"
printf '60000\n'      > "$h/temp3_input"
printf '87000000\n'   > "$h/power1_average"       # 87 W - proves ÷ 1e6
printf '1450\n'       > "$h/fan1_input"
printf '2100000000\n' > "$h/freq1_input"          # 2100 MHz
printf '1750000000\n' > "$h/freq2_input"

# card2: the sparse card (APU-like - proves the nil set)
d=$root/card2/device; h=$d/hwmon/hwmon5
mkdir -p "$h"
printf '7\n'          > "$d/gpu_busy_percent"
printf '536870912\n'  > "$d/mem_info_vram_used"
printf '4294967296\n' > "$d/mem_info_vram_total"
printf '134217728\n'  > "$d/mem_info_gtt_used"
printf '55000\n'      > "$h/temp1_input"
printf '15000000\n'   > "$h/power1_average"
printf '8000000000\n' > "$h/freq1_input"
# no temp2/temp3, no fan1, no freq2

# the traps detection must ignore
mkdir -p $root/card1-DP-1/device
printf 'connector\n' > $root/card1-DP-1/device/uevent
mkdir -p $root/card3/device
printf 'not-a-gpu\n' > $root/card3/device/uevent
