package peer

import (
	"testing"

	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/fixtures"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckpointCreationLastReached(t *testing.T) {
	// setup
	checkpoints := chaincfg.MainNetParams.Checkpoints
	testLogger := zerolog.Nop()

	cases := map[string]struct {
		checkpoints         []chaincfg.Checkpoint
		tipHeight           int32
		expectedLastReached bool
	}{
		"last checkpoint not reached at height 0": {
			checkpoints:         checkpoints,
			tipHeight:           0,
			expectedLastReached: false,
		},
		"last checkpoint not reached at height 810 000": {
			checkpoints:         checkpoints,
			tipHeight:           810000,
			expectedLastReached: false,
		},
		"last checkpoint reached at height 999 999 999": {
			checkpoints:         checkpoints,
			tipHeight:           999_999_999,
			expectedLastReached: true,
		},
		"last checkpoint reached if checkpoints list is empty": {
			checkpoints:         make([]chaincfg.Checkpoint, 0),
			tipHeight:           0,
			expectedLastReached: true,
		},
	}

	for name, params := range cases {
		t.Run(name, func(t *testing.T) {
			// given:
			chckPoint := newCheckpoint(params.checkpoints, params.tipHeight, &testLogger)

			// expect:
			assert.Equal(t, params.expectedLastReached, chckPoint.LastReached())
		})
	}

}

func TestCheckpointUseTip(t *testing.T) {
	// setup
	testLogger := zerolog.Nop()
	checkpoints := chaincfg.MainNetParams.Checkpoints
	lastCheckpoint := checkpoints[len(checkpoints)-1]

	cases := map[string]struct {
		checkpoints              []chaincfg.Checkpoint
		tipHeight                int32
		expectedCheckpointHeight int32
		expectedLastReached      bool
	}{
		"use height 0": {
			checkpoints:              checkpoints,
			tipHeight:                1,
			expectedCheckpointHeight: checkpoints[0].Height,
			expectedLastReached:      false,
		},
		"use height 11110": {
			checkpoints:              checkpoints,
			tipHeight:                11110,
			expectedCheckpointHeight: checkpoints[0].Height,
			expectedLastReached:      false,
		},
		"use height 11111": {
			checkpoints:              checkpoints,
			tipHeight:                11111,
			expectedCheckpointHeight: checkpoints[1].Height,
			expectedLastReached:      false,
		},
		"use height just before last checkpoint ": {
			checkpoints:              checkpoints,
			tipHeight:                lastCheckpoint.Height - 1,
			expectedCheckpointHeight: lastCheckpoint.Height,
			expectedLastReached:      false,
		},
		"use height at last checkpoint": {
			checkpoints:              checkpoints,
			tipHeight:                lastCheckpoint.Height,
			expectedCheckpointHeight: -1, // don't check that
			expectedLastReached:      true,
		},
		"use height just after last checkpoint": {
			checkpoints:              checkpoints,
			tipHeight:                lastCheckpoint.Height + 1,
			expectedCheckpointHeight: -1, // don't check that
			expectedLastReached:      true,
		},
	}

	for name, params := range cases {
		t.Run(name, func(t *testing.T) {
			// given:
			chckPoint := newCheckpoint(params.checkpoints, 0, &testLogger)

			// when:
			chckPoint.UseTip(params.tipHeight)
			// then:
			assert.Equal(t, params.expectedLastReached, chckPoint.LastReached())
			if !params.expectedLastReached {
				assert.Equal(t, params.expectedCheckpointHeight, chckPoint.Height())
			}
		})
	}
}

func TestCheckpointCurrentCheckpoint(t *testing.T) {
	// setup
	testLogger := zerolog.Nop()
	checkpoints := chaincfg.MainNetParams.Checkpoints
	lastCheckpoint := checkpoints[len(checkpoints)-1]

	cases := map[string]struct {
		header                   *domains.BlockHeader
		expectedLastReached      bool
		expectedCheckpointHeight int32
	}{
		"check header 11110 should keep first checkpoint": {
			header: &domains.BlockHeader{
				Height: 11110,
			},
			expectedCheckpointHeight: checkpoints[0].Height,
			expectedLastReached:      false,
		},
		"check header 11111 should set next checkpoint": {
			header: &domains.BlockHeader{
				Height: 11111,
				Hash:   *fixtures.HashOf("0000000069e244f73d78e8fd29ba2fd2ed618bd6fa2ee92559f542fdb26e7c1d"),
			},
			expectedCheckpointHeight: checkpoints[1].Height,
			expectedLastReached:      false,
		},
		"check header just before last checkpoint should keep last checkpoint as next checkpoint": {
			header: &domains.BlockHeader{
				Height: lastCheckpoint.Height - 1,
			},
			expectedCheckpointHeight: lastCheckpoint.Height,
			expectedLastReached:      false,
		},
		"check header at last checkpoint should set last checkpoint as next checkpoint": {
			header: &domains.BlockHeader{
				Height: lastCheckpoint.Height,
				Hash:   *lastCheckpoint.Hash,
			},
			expectedCheckpointHeight: 0,
			expectedLastReached:      true,
		},
	}

	for name, params := range cases {
		t.Run(name, func(t *testing.T) {
			// given:
			chckPoint := newCheckpoint(checkpoints, params.header.Height-1, &testLogger)

			// when:
			err := chckPoint.Check(params.header)

			// then:
			require.NoError(t, err)
			assert.Equal(t, params.expectedLastReached, chckPoint.LastReached())
			if !params.expectedLastReached {
				assert.Equal(t, params.expectedCheckpointHeight, chckPoint.Height())
			}
		})
	}

}

func TestCheckpointCheckSuccess(t *testing.T) {
	testLogger := zerolog.Nop()
	checkpoints := chaincfg.MainNetParams.Checkpoints
	lastCheckpoint := checkpoints[len(checkpoints)-1]
	cases := map[string]struct {
		checkpoints []chaincfg.Checkpoint
		header      *domains.BlockHeader
	}{
		"valid header reached nearest checkpoint": {
			checkpoints: checkpoints,
			header: &domains.BlockHeader{
				Height: 802000,
				Hash:   *fixtures.HashOf("000000000000000008f42d72af179115c35561d921f43829341967dcb8adbafd"),
			},
		},
		"header below nearest checkpoint": {
			checkpoints: checkpoints,
			header: &domains.BlockHeader{
				Height: 801999,
			},
		},
		"header reached last checkpoint": {
			checkpoints: checkpoints,
			header: &domains.BlockHeader{
				Height: lastCheckpoint.Height,
				Hash:   *lastCheckpoint.Hash,
			},
		},
		"header above last checkpoint": {
			checkpoints: checkpoints,
			header: &domains.BlockHeader{
				Height: lastCheckpoint.Height + 1,
			},
		},
		"checkpoint list is empty": {
			checkpoints: make([]chaincfg.Checkpoint, 0),
			header: &domains.BlockHeader{
				Height: 802000,
				Hash:   chaincfg.GenesisHash,
			},
		},
	}
	for name, params := range cases {
		t.Run(name, func(t *testing.T) {
			// given:
			chckPoint := newCheckpoint(params.checkpoints, params.header.Height-1, &testLogger)

			// expect:
			require.NoError(t, chckPoint.Check(params.header))
		})
	}
}

func TestCheckpointCheckFailure(t *testing.T) {
	testLogger := zerolog.Nop()
	checkpoints := chaincfg.MainNetParams.Checkpoints
	lastCheckpoint := checkpoints[len(checkpoints)-1]
	cases := map[string]struct {
		checkpoints    []chaincfg.Checkpoint
		previousHeight int32
		header         *domains.BlockHeader
		err            string
	}{
		"invalid header at checkpoint": {
			checkpoints:    checkpoints,
			previousHeight: 801999,
			header: &domains.BlockHeader{
				Height: 802000,
				Hash:   chaincfg.GenesisHash,
			},
			err: "corresponding checkpoint height does not match",
		},
		"unexpected header above the next checkpoint": {
			checkpoints:    checkpoints,
			previousHeight: 0,
			header: &domains.BlockHeader{
				Height: lastCheckpoint.Height,
				Hash:   *lastCheckpoint.Hash,
			},
			err: "unexpected header above next checkpoint height",
		},
	}
	for name, params := range cases {
		t.Run(name, func(t *testing.T) {
			// given:
			chckPoint := newCheckpoint(params.checkpoints, params.previousHeight, &testLogger)

			// when:
			err := chckPoint.Check(params.header)

			// then:
			require.Error(t, err)
			require.ErrorContains(t, err, params.err)
		})
	}
}
