package database_test

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"testing"

	"songloft/internal/database"
	"songloft/internal/database/testutil"
	"songloft/internal/models"
)

func TestPlaylistTextSortOnPooledConnections(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "sort.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	// Keep the first connection checked out while opening the second.
	for range 2 {
		conn, err := db.DB().Conn(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = conn.Close() })
		var got string
		if err := conn.QueryRowContext(ctx, "SELECT songloft_text_sort_key('  夜曲 Live ')").Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != "yequ live" {
			t.Fatalf("unexpected sort key %q", got)
		}
	}
}

// The clients paginate on the server and use song-ids to resume playback. Both
// queries must sort the entire playlist, including mixed Chinese/Latin text.
func TestPlaylistTextOrder(t *testing.T) {
	db := testutil.OpenMemoryDB(t)
	ctx := context.Background()
	playlist := &models.Playlist{Type: models.PlaylistTypeNormal, Name: "text order"}
	if err := db.PlaylistRepository().Create(ctx, playlist); err != nil {
		t.Fatal(err)
	}
	texts := []string{"夜曲", "beta", "愛你", "apple", "北京", "阿尔法", "APPLE", "爱你", "  月亮", "Zebra"}
	ids := make([]int64, len(texts))
	for i, text := range texts {
		song := &models.Song{
			Type: models.TypeLocal, Title: text, Artist: text, Album: text,
			FilePath: fmt.Sprintf("/text-order/%d.mp3", i), Duration: float64(i),
		}
		if err := db.SongRepository().Create(ctx, song); err != nil {
			t.Fatal(err)
		}
		ids[i] = song.ID
		if err := db.PlaylistSongRepository().AddSong(ctx, playlist.ID, song.ID, i+1); err != nil {
			t.Fatal(err)
		}
	}
	// Traditional/simplified homophones and case-only ties keep position order.
	textOrder := []int64{ids[5], ids[2], ids[7], ids[3], ids[6], ids[4], ids[1], ids[0], ids[8], ids[9]}
	for _, field := range []string{"title", "artist", "album", "duration", "position", "added_at"} {
		for _, direction := range []string{"asc", "desc"} {
			t.Run(field+"/"+direction, func(t *testing.T) {
				want := slices.Clone(textOrder)
				if field == "duration" || field == "position" || field == "added_at" {
					want = slices.Clone(ids)
				}
				if direction == "desc" {
					slices.Reverse(want)
				}
				gotIDs, err := db.PlaylistSongRepository().ListSongIDsOrdered(ctx, playlist.ID, field, direction)
				if err != nil {
					t.Fatal(err)
				}
				if !slices.Equal(gotIDs, want) {
					t.Errorf("song-ids = %v, want %v", gotIDs, want)
				}
				var pagedIDs []int64
				for offset := 0; offset < len(ids); offset += 3 {
					songs, err := db.PlaylistSongRepository().GetSongsFiltered(ctx, playlist.ID, database.PlaylistSongFilter{
						OrderBy: field, Order: direction, Limit: 3, Offset: offset,
					})
					if err != nil {
						t.Fatal(err)
					}
					for _, song := range songs {
						pagedIDs = append(pagedIDs, song.ID)
					}
				}
				if !slices.Equal(pagedIDs, want) {
					t.Errorf("paginated songs = %v, want %v", pagedIDs, want)
				}
			})
		}
	}
	filtered, err := db.PlaylistSongRepository().GetSongsFiltered(ctx, playlist.ID, database.PlaylistSongFilter{
		OrderBy: "title", Order: "asc", Keyword: "a", Limit: 2, Offset: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 2 || filtered[0].ID != ids[6] || filtered[1].ID != ids[1] {
		t.Fatalf("search must filter before sorting/pagination: %v", filtered)
	}
}
