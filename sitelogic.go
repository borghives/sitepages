package sitepages

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"git.mypierian.com/borghives/kosmos-go"
)

func NewBundle(name string) *Bundle {
	return &Bundle{
		Name: name,
	}
}

var MAX_BUNDLE_SIZE = 50

func AppendPage(ctx context.Context, bundle *Bundle, page Page) *Bundle {
	if len(bundle.Contents) >= MAX_BUNDLE_SIZE {
		kosmos.Record(ctx, bundle)
		newBundle := NewBundle(bundle.Name)
		newBundle.PreviousBundleId = bundle.GetID()
		bundle = newBundle
	}

	bundle.Contents = append(bundle.Contents, page.GetID())
	bundle.PageData = append(bundle.PageData, page)
	return bundle
}

func SaveSitePages(file string, pages []Page) error {
	// Open the file for writing
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(pages)
}

func GenerateMomentString(coolDown time.Duration) string {
	now := time.Now().UTC()
	return now.Add(coolDown).Format("2006-01-02 15:04")
}

func ParseMomentString(moment string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04", moment)
}

func LoadSitePages(site string) []Page {
	file, err := os.Open(site)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var retval []Page
	err = json.NewDecoder(file).Decode(&retval)
	if err != nil {
		log.Fatal(err)
	}
	return retval
}
