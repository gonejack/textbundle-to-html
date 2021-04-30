package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/djherbis/times"
	"github.com/gomarkdown/markdown"
)

type TextBundleToEpub struct {
	DefaultCover []byte
	Verbose      bool
}

func (t *TextBundleToEpub) Run(textBundles []string) (err error) {
	if len(textBundles) == 0 {
		return errors.New("no textbundle given")
	}

	for _, tb := range textBundles {
		err = t.process(tb)
		if err != nil {
			err = fmt.Errorf("parse %s failed: %s", tb, err)
			return
		}
	}

	return
}
func (t *TextBundleToEpub) process(textbundle string) (err error) {
	mdf, err := os.Open(filepath.Join(textbundle, "text.markdown"))
	if err != nil {
		return
	}
	defer mdf.Close()

	md, err := ioutil.ReadAll(mdf)
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(t.md2Html(md)))
	if err != nil {
		return
	}

	t.setBirthMeta(textbundle, doc)
	t.setTitle(textbundle, doc)
	t.changeImageRef(textbundle, doc)

	html, err := doc.Html()
	if err != nil {
		return
	}

	save := strings.TrimSuffix(textbundle, filepath.Ext(textbundle)) + ".html"
	err = ioutil.WriteFile(save, []byte(html), 0666)

	return
}
func (t *TextBundleToEpub) md2Html(md []byte) (html []byte) {
	return markdown.ToHTML(md, nil, nil)
}

func (t *TextBundleToEpub) setBirthMeta(textbundle string, doc *goquery.Document) {
	pubTime := time.Now()
	defer func() {
		meta := fmt.Sprintf(`<meta name="inostar:publish" content="%s">`, pubTime.Format(time.RFC1123Z))
		doc.Find("head").AppendHtml(meta)
	}()

	btime, err := getBirthTime(textbundle)
	if err == nil {
		pubTime = btime
		return
	}
	mtime, err := getModTime(textbundle)
	if err == nil {
		pubTime = mtime
		return
	}
}
func (t *TextBundleToEpub) setTitle(textbundle string, doc *goquery.Document) {
	if doc.Find("title").Length() == 0 {
		doc.Find("head").AppendHtml("<title></title>")
	}
	if doc.Find("title").Text() == "" {
		doc.Find("title").SetText(strings.TrimSuffix(filepath.Base(textbundle), filepath.Ext(textbundle)))
	}
}
func (t *TextBundleToEpub) changeImageRef(basedir string, doc *goquery.Document) {
	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		src, _ := img.Attr("src")
		switch {
		case src == "":
			return
		case strings.HasPrefix(src, "http"):
			return
		case strings.HasPrefix(src, "https"):
			return
		}

		// may be problems on Windows
		ref := filepath.Join(basedir, src)

		// be compatible with bear
		if _, err := os.Stat(ref); errors.Is(err, os.ErrNotExist) {
			src, _ = url.QueryUnescape(src)
			ref = filepath.Join(basedir, src)
		}

		img.SetAttr("src", ref)
	})
}

func getBirthTime(filepath string) (time.Time, error) {
	t, err := times.Stat(filepath)
	if err != nil {
		return time.Time{}, err
	}

	if !t.HasBirthTime() {
		return time.Time{}, errors.New("current OS does not support getting creation time of files")
	}

	return t.BirthTime(), nil
}

func getModTime(filepath string) (time.Time, error) {
	t, err := times.Stat(filepath)
	if err != nil {
		return time.Time{}, err
	}

	if !t.HasChangeTime() {
		return time.Time{}, errors.New("current OS does not support getting change time of files")
	}

	return t.ModTime(), nil
}
