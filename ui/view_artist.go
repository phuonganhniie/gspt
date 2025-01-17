package ui

import (
	"github.com/aditya-K2/gspt/spt"
	"github.com/zmb3/spotify/v2"
)

type ArtistView struct {
	*DefaultViewNone
	artistID  *spotify.ID
	topTracks []spotify.FullTrack
	albums    []spotify.SimpleAlbum
}

func NewArtistView() *ArtistView {
	a := &ArtistView{
		&DefaultViewNone{&defView{}},
		nil,
		[]spotify.FullTrack{},
		[]spotify.SimpleAlbum{},
	}
	return a
}

func (a *ArtistView) SetArtist(id *spotify.ID) {
	a.artistID = id
	a.topTracks = []spotify.FullTrack{}
	a.albums = []spotify.SimpleAlbum{}
	go func() {
		a.RefreshState()
	}()
}

func (a *ArtistView) RefreshState() {
	msg := SendNotificationWithChan("Loading Artist!")
	topTracks, err := spt.GetArtistTopTracks(*a.artistID)
	if err != nil {
		msg <- ("Error retrieving Artist Top Tracks: " + err.Error())
		return
	}
	a.topTracks = topTracks
	albums, err := spt.GetArtistAlbums(*a.artistID)
	if err != nil {
		msg <- ("Error retrieving Artist Albums: " + err.Error())
		return
	}
	a.albums = albums
	msg <- "Artist Loaded Succesfully!"
}

func (a *ArtistView) Content() func() [][]Content {
	return func() [][]Content {
		c := make([][]Content, 0)
		c = append(c, []Content{{"Artist Albums: ", NotSelectableStyle}})
		for _, v := range a.albums {
			c = append(c, []Content{
				{Content: v.Name, Style: AlbumStyle},
				{Content: artistName(v.Artists), Style: ArtistStyle},
				{Content: v.ReleaseDate, Style: TimeStyle},
			})
		}
		c = append(c, []Content{{"Artist Top Tracks:", NotSelectableStyle}})
		for _, v := range a.topTracks {
			c = append(c, []Content{
				{Content: v.Name, Style: TrackStyle},
				{Content: artistName(v.Artists), Style: ArtistStyle},
				{Content: v.Album.Name, Style: AlbumStyle},
			})
		}
		return c
	}
}

func (a *ArtistView) PlayEntry() {
	r, _ := Main.GetSelection()
	if r > 0 {
		if r < (len(a.albums) + 1) {
			if err := spt.PlayContext(a.albums[r-1].URI); err != nil {
				SendNotification(err.Error())
			}
		}
	}
}

func (a *ArtistView) OpenEntry() {
	r, _ := Main.GetSelection()
	if r > 0 {
		if r < (len(a.albums)+1) && len(a.albums) > 0 {
			albumView.SetAlbum(a.albums[r-1].Name, &a.albums[r-1].ID)
			SetCurrentView(albumView)
		} else if r != len(a.albums)+1 && len(a.topTracks) > 0 {
			if err := spt.PlaySong(a.topTracks[r-2-len(a.albums)].URI); err != nil {
				SendNotification(err.Error())
			}
		}
	}
}

func (a *ArtistView) Name() string { return "AlbumsView" }
