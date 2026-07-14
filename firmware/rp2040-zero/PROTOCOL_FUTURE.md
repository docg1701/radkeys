# Proposed firmware change — disambiguate GET_VERSION reply

## Problem
`readFirmwareVersion` reads 2 bytes. If a button is pressed during the
500ms version-read window, the button `[row, col]` IN report arrives
before the version reply and is misinterpreted as `[major, minor]`.
The current protocol uses no report ID to distinguish them (both use
`report_id=0`).

## Status

- **Host mitigation**: SHIPPED in v0.13.10. `internal/hid/reader_cgo.go`
  now retries `GET_VERSION` up to 3 times and accepts the first reply.
  This lowers the probability of a stray button press corrupting the
  version read, but it does not eliminate the ambiguity.

- **Firmware disambiguation**: PROPOSED, not implemented. This file
  documents the intended change for when Galvani has a RP2040-Zero
  prototype to flash and validate.

## Proposed firmware fix

Use a distinct report ID for `GET_VERSION` replies so the host can tell
apart normal button events from version replies.

### Changes in `diy.ino`

1. Define a new report ID constant near the top of the file:

```cpp
#define REPORT_ID_BUTTON  0 // normal 2-byte [row, col] IN reports
#define REPORT_ID_VERSION 1 // 2-byte [major, minor] GET_VERSION reply
```

2. Update the vendor report descriptor so it declares two INPUT reports:

```cpp
static const uint8_t desc_vendor[] = {
  TUD_HID_REPORT_DESC_GENERIC_INOUT_WITH_ID(REPORT_ID_BUTTON,  2),
  TUD_HID_REPORT_DESC_GENERIC_INOUT_WITH_ID(REPORT_ID_VERSION, 2)
};
```

> `TUD_HID_REPORT_DESC_GENERIC_INOUT_WITH_ID` may not exist in the
> version of TinyUSB bundled with the earlephilhower core. If it is
> unavailable, declare the descriptor manually or use a single report ID
> for all vendor IN reports and add a sentinel byte instead (see
> Alternative below).

3. Send button events with the button report ID:

```cpp
uint8_t report[2] = {uint8_t(r), uint8_t(c)};
usb_vendor.sendReport(REPORT_ID_BUTTON, report, sizeof(report));
```

4. Send the version reply with the version report ID:

```cpp
if (pending_version) {
  pending_version = 0;
  uint8_t ver[2] = {FW_VERSION_MAJOR, FW_VERSION_MINOR};
  usb_vendor.sendReport(REPORT_ID_VERSION, ver, sizeof(ver));
}
```

### Host changes that go with the firmware fix

1. In `internal/hid/reader_cgo.go`, read the report ID byte before the
   2-byte payload. The version read should only accept a reply whose
   first byte is `REPORT_ID_VERSION`. The event loop should drop any IN
   report whose first byte is not `REPORT_ID_BUTTON`.

2. Update the mock/fake device in tests to prepend the report ID byte
   so `readFirmwareVersion` and `loop()` continue to work in unit tests.

3. Update `firmware/rp2040-zero/PROTOCOL.md` to document the new
   report IDs.

## Alternative if report IDs are not available

If the TinyUSB/Adafruit_USBD_HID stack on the RP2040-Zero does not
support multiple report IDs on the vendor interface, use a sentinel byte
instead:

- Normal button IN report: `[0x00, row, col]` (3 bytes).
- Version IN reply: `[0x01, major, minor]` (3 bytes).

This changes the vendor interface report size from 2 to 3 bytes and
requires matching host changes. It is less elegant than report IDs but
achieves the same unambiguous framing.

## Decision gate

Do not implement either approach until Galvani can flash a RP2040-Zero
and confirm:

1. The device still enumerates as a composite HID device.
2. Normal button events reach the host unchanged.
3. `GET_VERSION` replies still reach the host.
4. The host-side version check reports the correct firmware version.
