package manifest

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"slices"
)

type MPD struct {
	Period struct {
		AdaptationSet struct {
			ContentProtection []struct{} `xml:"ContentProtection"`
			Representation    struct {
				Codecs string `xml:"codecs,attr"`
			}
		} `xml:"AdaptationSet"`
	} `xml:"Period"`
}

type JsonManifest struct {
	Codecs         string `json:"codecs"`
	EncryptionType string `json:"encryptionType,omitempty"`
}

type ManifestInfo struct {
	Codecs    string
	Encrypted bool
}

func GetInfoFromManifest(manifest string, mimeType string) (ManifestInfo, error) {
	manifestInfo := ManifestInfo{}

	var manifestData []byte
	manifestData, err := base64.StdEncoding.DecodeString(manifest)
	if err != nil {
		return manifestInfo, err
	}

	switch mimeType {
	case "application/dash+xml":
		mpd := MPD{}
		err = xml.Unmarshal(manifestData, &mpd)
		if err != nil {
			break
		}
		manifestInfo.Codecs = mpd.Period.AdaptationSet.Representation.Codecs
		manifestInfo.Encrypted = len(mpd.Period.AdaptationSet.ContentProtection) != 0
	case "application/vnd.tidal.bts", "application/vnd.tidal.emu":
		//fmt.Println(string(manifestData))
		manifest := JsonManifest{}
		err = json.Unmarshal(manifestData, &manifest)
		if err != nil {
			break
		}
		manifestInfo.Codecs = manifest.Codecs
		manifestInfo.Encrypted = !slices.Contains([]string{"NONE", ""}, manifest.EncryptionType)
	default:
		return manifestInfo, errors.New("unknown manifest MIME type")
	}

	return manifestInfo, err
}
