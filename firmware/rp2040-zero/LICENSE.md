# Firmware Licensing Notes

The `diy.ino` sketch and any original firmware files in this directory are
released under the same RadKeys Source-Available License v1.0 as the rest of
the project (see [LICENSE](/LICENSE)).

## Third-party components

When compiled, the resulting firmware binary includes third-party code from the
Arduino ecosystem:

- **Arduino-Pico core** by earlephilhower — licensed under the GNU Lesser General
  Public License v2.1 (LGPL v2.1).
- **Adafruit TinyUSB library** — licensed under the MIT License.

## What this means

- The RadKeys sketch source code remains under the RadKeys Source-Available
  License.
- The compiled binary is a combined work that includes LGPL code. Recipients of
  the binary have the rights granted by the LGPL regarding the LGPL-covered
  portions, including the right to obtain the corresponding source code of the
  Arduino-Pico core and to relink the binary with modified versions of that
  core.
- This does not require the RadKeys sketch source code to be released under the
  LGPL or any other open source license.

## Source code for LGPL components

The source code for the LGPL components is available from their respective
repositories:

- Arduino-Pico core: https://github.com/earlephilhower/arduino-pico
- Adafruit TinyUSB: https://github.com/adafruit/Adafruit_TinyUSB_Arduino
