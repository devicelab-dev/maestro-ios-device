# maestro-ios-device

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![macOS](https://img.shields.io/badge/platform-macOS-lightgrey.svg)](https://www.apple.com/macos/)
[![Maestro 2.x](https://img.shields.io/badge/Maestro-2.x-green.svg)](https://maestro.mobile.dev)

> 🚀 **Looking for more than just iOS device support?**
>
> **[maestro-runner](https://github.com/devicelab-dev/maestro-runner)** includes everything `maestro-ios-device` does — **3.6x faster, 14x less memory** (single binary, no JVM startup).
>
> - ✅ **Real iOS devices** — built-in, no patching needed
> - ✅ **Android emulators, simulators & real devices**
> - ✅ **Parallel execution** — dynamic work distribution across devices
> - ✅ **CI/CD ready** — instant cold starts, GitHub Actions / GitLab / Jenkins
> - ✅ **Cloud testing** — BrowserStack, Sauce Labs, LambdaTest via Appium
> - ✅ **Zero migration** — runs your existing Maestro YAML flows as-is
>
> ```bash
> curl -fsSL https://open.devicelab.dev/install/maestro-runner | bash
> maestro-runner --platform ios test flow.yaml
> ```
>
> 👉 **[Install maestro-runner →](https://github.com/devicelab-dev/maestro-runner)**

**Run Maestro UI tests on real iOS devices** — iPhone and iPad physical device testing with XCTest driver and automatic port forwarding.

> [!NOTE]
> This repo remains fully supported until Maestro adds official iOS physical device support.

> ⚠️ **Unofficial Community Tool**
> This is not affiliated with or endorsed by mobile.dev or the Maestro project.

> Built from [PR #2856](https://github.com/mobile-dev-inc/Maestro/pull/2856) — for source code, please refer to the [PR #2856](https://github.com/mobile-dev-inc/Maestro/pull/2856).

*Built by [DeviceLab](https://devicelab.dev) — stop renting devices you already own.*

## Table of Contents

- [Features](#what-it-does)
- [Compatibility](#compatibility)
- [Requirements](#requirements)
- [Installation](#installation)
- [Usage](#usage)
- [How It Works](#how-it-works)
- [Limitations](#limitations)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## What It Does

1. Patches your existing Maestro installation with real device support
2. Builds and runs the XCTest driver on your iOS device
3. Sets up port forwarding so Maestro can communicate with the device

## Compatibility

| Maestro Version | Status |
|-----------------|--------|
| 2.1.0 | ✅ Supported |
| 2.0.10 | ✅ Supported |
| 2.0.9 | ✅ Supported |
| Other 2.x | ❌ Not tested |
| 1.x | ❌ Not supported |

> **Note:** We build patches against specific Maestro releases. Using unsupported versions may cause issues.

## Requirements

- macOS
- [Maestro 2.x](https://maestro.mobile.dev) installed
- Xcode with command line tools
- Apple Developer account (free or paid)
- iOS device connected via USB

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/devicelab-dev/maestro-ios-device/main/setup.sh | bash
```

This will:

- Download the `maestro-ios-device` binary
- Download patched JARs and iOS runner
- Backup your current Maestro installation
- Patch Maestro with real device support

### Verify Installation

```bash
maestro-ios-device --version
```

## Usage

### 1. Start the Device Bridge

Keep this running in a terminal:

```bash
maestro-ios-device --team-id YOUR_TEAM_ID --device DEVICE_UDID
```

### 2. Run Maestro Tests

In another terminal:

```bash
maestro --driver-host-port 6001 --device DEVICE_UDID --app-file /path/to/app.ipa test flow.yaml
```

### Finding Your Team ID

```bash
# List available teams
security find-identity -v -p codesigning | grep "Developer"
```

Or in Xcode: **Xcode → Settings → Accounts → Select Team → Team ID**

### Finding Your Device UDID

```bash
# List connected devices
xcrun xctrace list devices
```

Or in Finder: **Select your iPhone → Click device name to reveal UDID**

## Options

| Flag | Description |
|------|-------------|
| `--team-id` | Apple Developer Team ID (required) |
| `--device` | Device UDID (required) |
| `--driver-host-port` | Local port (default: auto from 6001) |
| `--uninstall` | Restore original Maestro installation |
| `--version` | Show version |
| `--help` | Show help |

## How It Works

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────┐
│   Maestro   │────▶│ maestro-ios-device│────▶│ iOS Device  │
│  (patched)  │     │  (port forward)  │     │ (XCTest)    │
└─────────────┘     └──────────────────┘     └─────────────┘
     :6001                                        :22087
```

1. **maestro-ios-device** builds and installs the XCTest runner on your device
2. The runner starts an HTTP server on device port 22087
3. Port forwarding connects localhost:6001 → device:22087
4. Patched Maestro sends commands via `--driver-host-port 6001`

## Limitations

Some commands have limited support on real iOS devices due to iOS restrictions:

| Command | Status | Notes |
|---------|--------|-------|
| `clearState` | ✅ Works | Reinstalls app (requires `--app-file`) |
| `setLocation` | ⚠️ Limited | Requires additional setup |
| `addMedia` | ❌ Not supported | iOS restriction |

> **Note:** The `--app-file` flag is required for real device testing.

## Uninstallation

Restore your original Maestro installation:

```bash
maestro-ios-device --uninstall
```

Or manually:

```bash
cp ~/.maestro/backup/* ~/.maestro/lib/
```

## When to Stop Using This

This tool is temporary. Once [PR #2856](https://github.com/mobile-dev-inc/Maestro/pull/2856) is merged:

1. Run `maestro-ios-device --uninstall`
2. Update Maestro: `brew upgrade maestro`
3. Use official iOS device support

We'll update this README when official support lands.

## Troubleshooting

### "Certificate not trusted"

On your iOS device: **Settings → General → VPN & Device Management → Trust your developer certificate**

### Build fails

- Ensure Xcode command line tools are installed: `xcode-select --install`
- Open Xcode at least once to accept the license
- Check that your Apple Developer account is signed in

### Device not found

- Ensure device is connected via USB
- Trust the computer on your device when prompted
- Try `xcrun xctrace list devices` to verify connection

### Port already in use

```bash
# Find process using port 6001
lsof -i :6001

# Use a different port
maestro-ios-device --team-id YOUR_TEAM_ID --device DEVICE_UDID --driver-host-port 6002
```

### XCTest runner crashes

- Ensure your device is running iOS 15+
- Check Xcode logs: **Window → Devices and Simulators → View Device Logs**

## Contributing

Issues and PRs welcome at [GitHub](https://github.com/devicelab-dev/maestro-ios-device/issues).

## License

Apache 2.0 (same as Maestro)

## Disclaimer

This project is not affiliated with, endorsed by, or connected to mobile.dev or the official Maestro project.

This tool patches your existing Maestro installation to add functionality not yet available in the official release.

**Use at your own risk.** We recommend switching to official Maestro once iOS device support is released.

---

[Report Issues](https://github.com/devicelab-dev/maestro-ios-device/issues)
