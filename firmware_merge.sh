#!/bin/bash
esptool.py --chip esp32c6 merge_bin -o firmware-merged.bin \
  0x0      build/bootloader/bootloader.bin \
  0x8000   build/partition_table/partition-table.bin \
  0xF000   build/ota_data_initial.bin \
  0x20000  build/wifi_station.bin

esptool.py --chip esp32c6 write_flash 0x0 firmware-merged.bin
