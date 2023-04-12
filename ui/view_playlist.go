package ui

import (
	"fmt"

	"github.com/aditya-K2/gspt/spt"
	"github.com/gdamore/tcell/v2"
	"github.com/zmb3/spotify/v2"
)

type PlaylistView struct {
	*DefaultView
	currentPlaylist         *spotify.SimplePlaylist
	currentUserFullPlaylist *spt.Playlist
	I                       *interactiveView
}

func (p *PlaylistView) SetPlaylist(pl *spotify.SimplePlaylist) {
	p.currentPlaylist = pl
	p.currentUserFullPlaylist = nil
}

func (p *PlaylistView) Content() func() [][]Content {
	return func() [][]Content {
		c := make([][]Content, 0)
		if p.currentPlaylist != nil {
			if p.currentUserFullPlaylist == nil {
				msg := SendNotificationWithChan(fmt.Sprintf("Loading %s....", p.currentPlaylist.Name))
				pf, err := spt.GetPlaylist(p.currentPlaylist.ID, func(s bool, e error) {
					go func() {
						if !s {
							msg <- e.Error()
						} else {
							msg <- "Playlist Loaded Succesfully"
						}
					}()
				})
				if err != nil {
					SendNotification(fmt.Sprintf("Error Retrieving %s", p.currentPlaylist.Name))
					return [][]Content{}
				}
				p.currentUserFullPlaylist = pf
			}
			for _, v := range *(*p.currentUserFullPlaylist).Tracks {
				c = append(c, []Content{
					{Content: v.Track.Name, Style: TrackStyle},
					{Content: v.Track.Artists[0].Name, Style: ArtistStyle},
					{Content: v.Track.Album.Name, Style: AlbumStyle},
				})
			}
		}
		return c
	}
}

func (p *PlaylistView) ContextHandler() func(start, end, sel int) {
	return func(start, end, sel int) {
		// Assuming that there are no external effects on the user's playlists
		// (i.e Any Creation or Deletion of Playlists while the context Menu is
		// open
		userPlaylists, err := spt.CurrentUserPlaylists(func(s bool, err error) {})
		if err != nil {
			SendNotification("Error Retrieving User Playlists")
			return
		}
		tracks := make([]spotify.ID, 0)
		for k := start; k <= end; k++ {
			tracks = append(tracks, (*(*p.currentUserFullPlaylist).Tracks)[k].Track.ID)
		}
		aerr := spt.AddTracksToPlaylist((*userPlaylists)[sel].ID, tracks...)
		if aerr != nil {
			SendNotification(aerr.Error())
			return
		} else {
			SendNotification(fmt.Sprintf("Added %d tracks to %s", len(tracks), (*userPlaylists)[sel].Name))
		}
	}
}

func (p *PlaylistView) ExternalInputCapture() func(e *tcell.EventKey) *tcell.EventKey {
	return func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEnter {
			r, _ := Ui.Main.Table.GetSelection()
			if err := spt.PlaySongWithContext(&p.currentPlaylist.URI, r); err != nil {
				SendNotification(err.Error())
			}
		}
		return e
	}
}

func (p *PlaylistView) Name() string { return "PlaylistView" }
