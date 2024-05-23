package peer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPeersCollection_AddPeer(t *testing.T) {
	t.Run("add peer", func(t *testing.T) {
		// given
		sut := NewPeersCollection(3)

		// when
		err := sut.AddPeer(&Peer{})

		// then
		require.NoError(t, err)
	})

	t.Run("no space for new item", func(t *testing.T) {
		// given
		sut := NewPeersCollection(1)
		sut.AddPeer(&Peer{})

		// when
		err := sut.AddPeer(&Peer{})

		// then
		require.Error(t, err)
	})
}

func TestPeersCollection_RmPeer(t *testing.T) {
	t.Run("peer doesn't exist in collection", func(t *testing.T) {
		// given
		sut := NewPeersCollection(3)
		sut.AddPeer(&Peer{})

		// when
		sut.RmPeer(&Peer{})

		// then
		require.Len(t, sut.Enumerate(), 1)
	})

	t.Run("peer exists in collection", func(t *testing.T) {
		// given
		sut := NewPeersCollection(3)
		p1 := &Peer{}
		sut.AddPeer(p1)

		// when
		sut.RmPeer(p1)

		// then
		require.Empty(t, sut.Enumerate())
	})
}

func TestPeersCollection_Space(t *testing.T) {
	t.Run("empty collection", func(t *testing.T) {
		// given
		const size = uint(3)
		sut := NewPeersCollection(size)

		// when
		space := sut.Space()

		// then
		require.Equal(t, size, space)
	})

	t.Run("grow only", func(t *testing.T) {
		// given
		const size = uint(3)
		sut := NewPeersCollection(size)

		// when
		sut.AddPeer(&Peer{})
		space := sut.Space()

		// then
		require.Equal(t, size-1, space)
	})

	t.Run("grow and shrink", func(t *testing.T) {
		// given
		const size = uint(3)
		sut := NewPeersCollection(size)
		p1 := &Peer{}
		p2 := &Peer{}
		p3 := &Peer{}

		// when
		sut.AddPeer(p1)
		sut.AddPeer(p2)
		sut.AddPeer(p3)

		sut.RmPeer(p1)
		sut.RmPeer(p2)

		space := sut.Space()

		// then
		require.Equal(t, size-1, space)

	})

}

func TestPeersCollection_Enumerate(t *testing.T) {
	t.Run("empty collection", func(t *testing.T) {
		// given
		sut := NewPeersCollection(3)

		// when
		enumerated := sut.Enumerate()

		// then
		require.Empty(t, enumerated)

	})

	t.Run("grow only", func(t *testing.T) {
		// given
		sut := NewPeersCollection(3)

		// when
		sut.AddPeer(&Peer{})
		enumerated := sut.Enumerate()

		// then
		require.Len(t, enumerated, 1)

		for _, p := range enumerated {
			require.NotNil(t, p)
		}
	})

	t.Run("grow and shrink", func(t *testing.T) {
		// given
		sut := NewPeersCollection(3)
		p1 := &Peer{}
		p2 := &Peer{}
		p3 := &Peer{}

		// when
		sut.AddPeer(p1)
		sut.AddPeer(p2)
		sut.AddPeer(p3)

		sut.RmPeer(p1)
		sut.RmPeer(p2)

		enumerated := sut.Enumerate()

		// then
		require.Len(t, enumerated, 1)

		for _, p := range enumerated {
			require.NotNil(t, p)
		}

	})
}
