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

### Real APIs (confirmed from TinyUSB upstream)

- **Report descriptor macro**: `TUD_HID_REPORT_DESC_GENERIC_INOUT(report_size, ...)`
  takes variadic args. Pass `HID_REPORT_ID(n)` to add a report ID:

  ```c
  // Single report, no ID (current usage)
  TUD_HID_REPORT_DESC_GENERIC_INOUT(2)

  // Single report with ID
  TUD_HID_REPORT_DESC_GENERIC_INOUT(2, HID_REPORT_ID(REPORT_ID_BUTTON))

  // Two separate vendor collections, each with its own report ID
  TUD_HID_REPORT_DESC_GENERIC_INOUT(2, HID_REPORT_ID(REPORT_ID_BUTTON)),
  TUD_HID_REPORT_DESC_GENERIC_INOUT(2, HID_REPORT_ID(REPORT_ID_VERSION))
  ```

  Source: <https://github.com/hathach/tinyusb/blob/master/src/class/hid/hid_device.h#L423>

- **sendReport**: `usb_vendor.sendReport(report_id, buffer, len)` (Adafruit_USBD_HID
  wrapper). The `report_id` parameter is real and passed through to
  `tud_hid_n_report(instance, report_id, report, len)`.

  Source: <https://github.com/hathach/tinyusb/blob/master/src/class/hid/hid_device.h#L71>

> There is **no** `TUD_HID_REPORT_DESC_GENERIC_INOUT_WITH_ID` macro — that was
> invented by an earlier draft. The correct macro is `GENERIC_INOUT` with
> `HID_REPORT_ID(n)` in the variadic position.

### Changes in `diy.ino`

1. Define two report ID constants near the top of the file:

```cpp
#define REPORT_ID_BUTTON  0 // 2-byte [row, col] button IN reports
#define REPORT_ID_VERSION 1 // 2-byte [major, minor] GET_VERSION reply
```

2. Change the vendor report descriptor to declare two separate
   application collections, each with its own report ID. The first
   covers button events (IN) and host commands (OUT); the second
   covers only the version reply (IN — its OUT goes unused):

```cpp
static const uint8_t desc_vendor[] = {
  TUD_HID_REPORT_DESC_GENERIC_INOUT(2, HID_REPORT_ID(REPORT_ID_BUTTON)),
  TUD_HID_REPORT_DESC_GENERIC_INOUT(2, HID_REPORT_ID(REPORT_ID_VERSION))
};
```

3. Replace the version reply in `loop()` to use the version report ID:

```cpp
if (pending_version) {
  pending_version = 0;
  uint8_t ver[2] = {FW_VERSION_MAJOR, FW_VERSION_MINOR};
  usb_vendor.sendReport(REPORT_ID_VERSION, ver, sizeof(ver));
}
```

Button events already use `usb_vendor.sendReport(0, report, sizeof(report))` —
change the `0` to `REPORT_ID_BUTTON`.

### Host changes that go with the firmware fix

1. In `internal/hid/reader_cgo.go`, the `readFirmwareVersion` function
   must now expect a 3-byte reply (report_id + payload): read 3 bytes,
   verify the first byte is `REPORT_ID_VERSION`, then return `buf[1]`
   and `buf[2]` as major/minor.

2. In the event loop, skip any IN report whose first byte is not
   `REPORT_ID_BUTTON` (i.e., drop stray version replies that arrive
   between button events).

3. Update `internal/hid/mock.go` and any fake device implementations to
   prepend the report ID byte in both button and version replies.

4. Update `firmware/rp2040-zero/PROTOCOL.md` to document the new
   report IDs.

## Alternative: sentinel byte (if report IDs don't work)

If the TinyUSB/earlephilhower stack rejects two GENERIC_INOUT
collections on the same vendor page, use a sentinel byte instead:

- Button IN report: `[REPORT_ID_BUTTON, row, col]` (3 bytes)
- Version IN reply: `[REPORT_ID_VERSION, major, minor]` (3 bytes)
- Vendor descriptor stays as a single `TUD_HID_REPORT_DESC_GENERIC_INOUT(3)`
  (no report ID, but report size changes from 2 to 3).
- The host always reads 3 bytes and dispatches by the first byte.
- Downside: report size grows from 2 to 3 bytes across all button events.

## Decision gate

Do not implement either approach until Galvani can flash a RP2040-Zero
and confirm:

1. The device still enumerates as a composite HID device.
2. Normal button events reach the host (with or without report ID).
3. `GET_VERSION` replies reach the host with the correct report ID.
4. The host-side version check reports the correct firmware version.
5. The composite keyboard interface (Interface B) is unaffected.
