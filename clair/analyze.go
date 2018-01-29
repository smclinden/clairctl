package clair

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coreos/clair/api/v1"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/docker/reference"
)

//Analyze return Clair Image analysis
func Analyze(image reference.NamedTagged, manifest distribution.Manifest) ImageAnalysis {
	layers, err := newLayering(image)
	if err != nil {
		log.Fatalf("cannot parse manifest")
		return ImageAnalysis{}
	}

	switch manifest.(type) {
	case schema1.SignedManifest:

		for _, l := range manifest.(schema1.SignedManifest).FSLayers {
			layers.digests = append(layers.digests, l.BlobSum.String())
		}
		return layers.analyzeAll()
	case *schema1.SignedManifest:
		for _, l := range manifest.(*schema1.SignedManifest).FSLayers {
			layers.digests = append(layers.digests, l.BlobSum.String())
		}
		return layers.analyzeAll()
	case schema2.DeserializedManifest:
		log.Debugf("json: %v", image)
		for _, l := range manifest.(schema2.DeserializedManifest).Layers {
			layers.digests = append(layers.digests, l.Digest.String())
		}
		return layers.analyzeAll()
	case *schema2.DeserializedManifest:
		log.Debugf("json: %v", image)
		for _, l := range manifest.(*schema2.DeserializedManifest).Layers {
			layers.digests = append(layers.digests, l.Digest.String())
		}
		return layers.analyzeAll()
	default:
		log.Fatalf("Unsupported Schema version.")
		return ImageAnalysis{}
	}
}

func analyzeLayer(id string) (v1.LayerEnvelope, error) {

	lURI := fmt.Sprintf("%v/layers/%v?vulnerabilities", uri, id)
	log.Debugf("GETing url analysis: %v", lURI)

	request, err := http.NewRequest("GET", lURI, nil)
	if err != nil {
		return v1.LayerEnvelope{}, err
	}
	SetRequestHeaders(request)

	response, err := (&http.Client{}).Do(request)

	if err != nil {
		return v1.LayerEnvelope{}, fmt.Errorf("analysing layer %v: %v", id, err)
	}
	defer response.Body.Close()

	var analysis v1.LayerEnvelope
	err = json.NewDecoder(response.Body).Decode(&analysis)
	if err != nil {
		return v1.LayerEnvelope{}, fmt.Errorf("reading layer analysis: %v", err)
	}
	if response.StatusCode != 200 {
		log.Infof("If you are seeing 400 errors on a local scan it likely means that the url Clair is using to callback to the clairCtl utility does not match temp location.")
		log.Infof("Review the Clair logs for the failing url, and make sure to reconcile the path after /local with the actual location clairCtl stores the image.")
		log.Infof("Make sure to use the --log-level DEBUG flag to view the information required")
		return v1.LayerEnvelope{}, fmt.Errorf("receiving http error: %d", response.StatusCode)
	}

	return analysis, nil
}
