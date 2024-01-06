package api

import (
	"fmt"
	"net/http"
	"net/url"
)

type TrackPlaybackInfo struct {
	TrackID              int     `json:"trackId"`
	AssetPresentation    string  `json:"assetPresentation"`
	AudioMode            string  `json:"audioMode"`
	AudioQuality         string  `json:"audioQuality"`
	StreamingSessionID   string  `json:"streamingSessionId"`
	LicenseSecurityToken string  `json:"licenseSecurityToken"`
	ManifestMimeType     string  `json:"manifestMimeType"`
	ManifestHash         string  `json:"manifestHash"`
	Manifest             string  `json:"manifest"`
	AlbumReplayGain      float64 `json:"albumReplayGain"`
	AlbumPeakAmplitude   float64 `json:"albumPeakAmplitude"`
	TrackReplayGain      float64 `json:"trackReplayGain"`
	TrackPeakAmplitude   float64 `json:"trackPeakAmplitude"`
}

func (a *TidalApi) GetTrackPlaybackInfo(id int, quality string, playbackMode string) (resp *TrackPlaybackInfo, err error) {
	params := url.Values{}
	params.Set("audioquality", quality)
	params.Set("playbackmode", playbackMode)
	params.Set("assetpresentation", "FULL")

	resp = &TrackPlaybackInfo{}

	err = a.call(http.MethodGet, fmt.Sprintf("tracks/%d/playbackinfo", id), params, resp)
	return
}

type EntityListResp struct {
	Limit              int `json:"limit"`
	Offset             int `json:"offset"`
	TotalNumberOfItems int `json:"totalNumberOfItems"`
	Items              []struct {
		ID             int    `json:"id"`
		Title          string `json:"title"`
		Duration       int    `json:"duration"`
		AllowStreaming bool   `json:"allowStreaming"`
		//TrackNumber    int    `json:"trackNumber"`
		Explicit      bool `json:"explicit"`
		MediaMetadata struct {
			Tags []string `json:"tags"`
		} `json:"mediaMetadata"`
		Artist struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"artist"`
		Album struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"album"`
	} `json:"items"`
}

func (a *TidalApi) GetTracksByIsrc(isrc string) (resp *EntityListResp, err error) {
	params := url.Values{}
	params.Set("isrc", isrc)
	params.Set("countryCode", a.countryCode)

	resp = &EntityListResp{}

	err = a.call(http.MethodGet, "tracks", params, resp)
	return
}

func (a *TidalApi) GetAlbumsByUpc(upc string) (resp *EntityListResp, err error) {
	params := url.Values{}
	params.Set("upc", upc)
	params.Set("countryCode", a.countryCode)

	resp = &EntityListResp{}

	err = a.call(http.MethodGet, "albums", params, resp)
	return
}
