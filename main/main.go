// (c) 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/djt-labs/paaro/app/process"
	"github.com/djt-labs/paaro/config"
	"github.com/djt-labs/paaro/utils"
	"github.com/djt-labs/paaro/utils/logging"
	"github.com/djt-labs/paaro/version"
)

// main is the entry point to Paaro.
func main() {
	fs := config.BuildFlagSet()
	v, err := config.BuildViper(fs, os.Args[1:])
	if err != nil {
		fmt.Printf("couldn't configure flags: %s\n", err)
		os.Exit(1)
	}

	processConfig, err := config.GetProcessConfig(v)
	if err != nil {
		fmt.Printf("couldn't load process config: %s\n", err)
		os.Exit(1)
	}

	if processConfig.DisplayVersionAndExit {
		fmt.Print(version.String)
		os.Exit(0)
	}

	nodeConfig, err := config.GetNodeConfig(v, processConfig.BuildDir)
	if err != nil {
		fmt.Printf("couldn't load node config: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(process.Header)

	// Set the log directory for this process
	logFactory := logging.NewFactory(nodeConfig.LoggingConfig)

	log, err := logFactory.Make("daemon")
	if err != nil {
		logFactory.Close()

		fmt.Printf("starting logger failed with: %s\n", err)
		os.Exit(1)
	}

	log.Info("using build directory at path '%s'", processConfig.BuildDir)

	nodeManager := newNodeManager(processConfig.BuildDir, log)
	utils.HandleSignals(
		func(os.Signal) {
			// SIGINT and SIGTERM cause all running nodes to stop
			nodeManager.shutdown()
		},
		syscall.SIGINT, syscall.SIGTERM,
	)

	// Run normally
	exitCode, err := nodeManager.runNormal()
	if err != nil {
		log.Error("running node returned error: %s", err)
	} else {
		log.Debug("node returned exit code %d", exitCode)
	}

	nodeManager.shutdown() // make sure all the nodes are stopped

	logFactory.Close()
	os.Exit(exitCode)
}
