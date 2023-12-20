package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/bitcoin-sv/pulse/internal/chaincfg"
	"github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"
)

// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
func cleanAndExpandPath(path string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		homeDir := filepath.Dir(DefaultHomeDir)
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but they variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

func maxUint32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

// parseCheckpoints checks the checkpoint strings for valid syntax
// ('<height>:<hash>') and parses them to chaincfg.Checkpoint instances.
func parseCheckpoints(checkpointStrings []string) ([]chaincfg.Checkpoint, error) {
	if len(checkpointStrings) == 0 {
		return nil, nil
	}
	checkpoints := make([]chaincfg.Checkpoint, len(checkpointStrings))
	for i, cpString := range checkpointStrings {
		checkpoint, err := newCheckpointFromStr(cpString)
		if err != nil {
			return nil, err
		}
		checkpoints[i] = checkpoint
	}
	return checkpoints, nil
}

// newCheckpointFromStr parses checkpoints in the '<height>:<hash>' format.
func newCheckpointFromStr(checkpoint string) (chaincfg.Checkpoint, error) {
	parts := strings.Split(checkpoint, ":")
	if len(parts) != 2 {
		return chaincfg.Checkpoint{}, fmt.Errorf("unable to parse "+
			"checkpoint %q -- use the syntax <height>:<hash>",
			checkpoint)
	}

	height, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return chaincfg.Checkpoint{}, fmt.Errorf("unable to parse "+
			"checkpoint %q due to malformed height", checkpoint)
	}

	if len(parts[1]) == 0 {
		return chaincfg.Checkpoint{}, fmt.Errorf("unable to parse "+
			"checkpoint %q due to missing hash", checkpoint)
	}
	hash, err := chainhash.NewHashFromStr(parts[1])
	if err != nil {
		return chaincfg.Checkpoint{}, fmt.Errorf("unable to parse "+
			"checkpoint %q due to malformed hash", checkpoint)
	}

	return chaincfg.Checkpoint{
		Height: int32(height),
		Hash:   hash,
	}, nil
}

// checkpointSorter implements sort.Interface to allow a slice of checkpoints to
// be sorted.
type checkpointSorter []chaincfg.Checkpoint

// Len returns the number of checkpoints in the slice.  It is part of the
// sort.Interface implementation.
func (s checkpointSorter) Len() int {
	return len(s)
}

// Swap swaps the checkpoints at the passed indices.  It is part of the
// sort.Interface implementation.
func (s checkpointSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less returns whether the checkpoint with index i should sort before the
// checkpoint with index j.  It is part of the sort.Interface implementation.
func (s checkpointSorter) Less(i, j int) bool {
	return s[i].Height < s[j].Height
}

// mergeCheckpoints returns two slices of checkpoints merged into one slice
// such that the checkpoints are sorted by height.  In the case the additional
// checkpoints contain a checkpoint with the same height as a checkpoint in the
// default checkpoints, the additional checkpoint will take precedence and
// overwrite the default one.
func mergeCheckpoints(defaultCheckpoints, additional []chaincfg.Checkpoint) []chaincfg.Checkpoint {
	// Create a map of the additional checkpoints to remove duplicates while
	// leaving the most recently-specified checkpoint.
	extra := make(map[int32]chaincfg.Checkpoint)
	for _, checkpoint := range additional {
		extra[checkpoint.Height] = checkpoint
	}

	// Add all default checkpoints that do not have an override in the
	// additional checkpoints.
	numDefault := len(defaultCheckpoints)
	checkpoints := make([]chaincfg.Checkpoint, 0, numDefault+len(extra))
	for _, checkpoint := range defaultCheckpoints {
		if _, exists := extra[checkpoint.Height]; !exists {
			checkpoints = append(checkpoints, checkpoint)
		}
	}

	// Append the additional checkpoints and return the sorted results.
	for _, checkpoint := range extra {
		checkpoints = append(checkpoints, checkpoint)
	}
	sort.Sort(checkpointSorter(checkpoints))
	return checkpoints
}

// BsvdLookup resolves the IP of the given host using the correct DNS lookup
// function depending on the configuration options.  For example, addresses will
// be resolved using tor when the --proxy flag was specified unless --noonion
// was also specified in which case the normal system DNS resolver will be used.
//
// Any attempt to resolve a tor address (.onion) will return an error since they
// are not intended to be resolved outside of the tor proxy.
func (c *P2PConfig) BsvdLookup(host string) ([]net.IP, error) {
	if strings.HasSuffix(host, ".onion") {
		return nil, fmt.Errorf("attempt to resolve tor address %s", host)
	}

	return c.lookup(host)
}

// BsvdDial connects to the address on the named network using the appropriate
// dial function depending on the address and configuration options.  For
// example, .onion addresses will be dialed using the onion specific proxy if
// one was specified, but will otherwise use the normal dial function (which
// could itself use a proxy or not).
func (c *P2PConfig) BsvdDial(addr net.Addr) (net.Conn, error) {
	if strings.Contains(addr.String(), ".onion:") {
		return nil, fmt.Errorf("no possibility to use onion addresses: %v", addr)
	}
	return c.dial(addr.Network(), addr.String(), DefaultConnectTimeout)
}

// getWorkingDirectory will get the current working directory.
func getWorkingDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}
