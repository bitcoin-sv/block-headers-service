package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// P2PConfig struct for p2pconfig.

// Validate validates config and sets some parameters based on the config provided.
func (c *P2PConfig) Validate() (err error) {
	funcName := "validateConfig"

	// Append the network type to the log directory so it is "namespaced"
	// per network in the same fashion as the data directory.
	c.LogDir = cleanAndExpandPath(c.LogDir)
	c.LogDir = filepath.Join(c.LogDir, netName(ActiveNetParams))

	// Don't allow ban durations that are too short.
	if c.BanDuration < time.Second {
		str := "%s: The banduration option may not be less than 1s -- parsed [%v]"
		err := fmt.Errorf(str, funcName, c.BanDuration)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	// Excessive blocksize cannot be set less than the default but it can be higher.
	c.ExcessiveBlockSize = maxUint32(c.ExcessiveBlockSize, DefaultExcessiveBlockSize)

	c.dial = net.DialTimeout
	c.lookup = net.LookupIP

	if err != nil {
		str := "%s: Error parsing checkpoints: %v"
		err := fmt.Errorf(str, funcName, err)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	if !c.DisableCheckpoints {
		c.Checkpoints = ActiveNetParams.Checkpoints
	}

	return nil
}
