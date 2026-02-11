package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/anthropics/maestro-ios-device/internal/device"
	"github.com/anthropics/maestro-ios-device/internal/maestro"
	"github.com/anthropics/maestro-ios-device/internal/portforward"
	"github.com/anthropics/maestro-ios-device/internal/runner"
	"github.com/anthropics/maestro-ios-device/internal/utils"
)

var version = "dev" // set via -ldflags

func fatal(format string, args ...any) {
	fmt.Printf("❌ "+format+"\n", args...)
	os.Exit(1)
}

func printBanner() {
	fmt.Printf("maestro-ios-device %s\n", version)
	fmt.Println("  🚀 3.6x faster, real iOS device support, runs locally or on any Appium cloud,")
	fmt.Println("  true parallel execution, no paywall. Fixes 78% of Maestro's top issues. Same YAML.")
	fmt.Println("  https://github.com/devicelab-dev/maestro-runner")
	fmt.Println("  Built by DeviceLab — https://devicelab.dev")
	fmt.Println()
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "setup" {
		printBanner()
		if err := maestro.RunSetup(); err != nil {
			fatal("Setup failed: %s", err)
		}
		return
	}
	run()
}

func run() {
	teamID := flag.String("team-id", "", "Apple Developer Team ID (required)")
	deviceUDID := flag.String("device", "", "Target device UDID (required)")
	port := flag.Int("driver-host-port", 0, "Local port (default: auto-assign from 6001)")
	showVersion := flag.Bool("version", false, "Show version")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *showVersion {
		fmt.Printf("maestro-ios-device %s\n", version)
		fmt.Println("  🚀 3.6x faster, real iOS device support, runs locally or on any Appium cloud,")
		fmt.Println("  true parallel execution, no paywall. Fixes 78% of Maestro's top issues. Same YAML.")
		fmt.Println("  https://github.com/devicelab-dev/maestro-runner")
		fmt.Println("  Built by DeviceLab — https://devicelab.dev")
		return
	}

	printBanner()

	if *help {
		printUsage()
		return
	}

	if *teamID == "" || *deviceUDID == "" {
		printUsage()
		os.Exit(1)
	}

	if ok, _ := maestro.IsPatched(); !ok {
		fatal("Maestro not patched. Run: maestro-ios-device setup")
	}

	localPort, err := utils.ResolvePort(*port)
	if err != nil {
		fatal("%s", err)
	}

	dev, err := device.Get(*deviceUDID)
	if err != nil {
		fatal("%s", err)
	}
	fmt.Printf("📱 %s (%s) - iOS %s\n\n", dev.Name, dev.Serial, dev.OSVersion)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Build & start runner
	r := runner.New(*deviceUDID, *teamID)
	defer r.Cleanup()

	if err := r.Build(ctx); err != nil {
		fatal("Build failed: %s", err)
	}

	if err := r.Start(ctx); err != nil {
		fatal("Start failed: %s", err)
	}

	pf := portforward.New(dev.Entry, uint16(localPort), runner.DevicePort)
	defer pf.Stop()

	if err := pf.Start(); err != nil {
		fatal("Port forward failed: %s", err)
	}

	if err := pf.Verify(); err != nil {
		fatal("%s", err)
	}

	fmt.Println()
	fmt.Println("✅ Ready! Run:")
	fmt.Printf("   maestro --driver-host-port %d --device %s --app-file /path/to/app.ipa test flow.yaml\n\n", localPort, *deviceUDID)
	fmt.Println("Press Ctrl+C to stop.")

	<-sigChan
	fmt.Println("\n🛑 Stopping...")
}

func printUsage() {
	fmt.Println(`maestro-ios-device - Run Maestro tests on real iOS devices

🚀 We built maestro-runner from scratch — 3.6x faster, real iOS device support,
   runs locally or on any Appium cloud, true parallel execution, no paywall.
   Fixes 78% of Maestro's top issues. Same YAML.
   https://github.com/devicelab-dev/maestro-runner

Usage:
  maestro-ios-device --team-id TEAM_ID --device UDID [options]

Required:
  --team-id       Apple Developer Team ID
  --device        iOS device UDID

Options:
  --driver-host-port   Local port for Maestro connection (default: 6001)
  --version            Show version
  --help               Show this help

Examples:
  maestro-ios-device --team-id ABC123XYZ --device 00008030-001234567890

  Then run tests:
  maestro --driver-host-port 6001 --device 00008030-001234567890 test flow.yaml

Finding your Team ID:
  security find-identity -v -p codesigning | grep "Developer"

Finding your Device UDID:
  xcrun xctrace list devices

Docs: https://github.com/devicelab-dev/maestro-ios-device
Built by DeviceLab — https://devicelab.dev`)
}
