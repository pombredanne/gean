package geanlib

import (
	"reflect"
	"testing"

	"github.com/geego/gean/app/deps"
	"github.com/geego/gean/app/tpl"
	"github.com/govenue/require"
)

const sitemapTemplate = `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  {{ range .Data.Pages }}
  <url>
    <loc>{{ .Permalink }}</loc>{{ if not .Lastmod.IsZero }}
    <lastmod>{{ safeHTML ( .Lastmod.Format "2006-01-02T15:04:05-07:00" ) }}</lastmod>{{ end }}{{ with .Sitemap.ChangeFreq }}
    <changefreq>{{ . }}</changefreq>{{ end }}{{ if ge .Sitemap.Priority 0.0 }}
    <priority>{{ .Sitemap.Priority }}</priority>{{ end }}
  </url>
  {{ end }}
</urlset>`

func TestSitemapOutput(t *testing.T) {
	t.Parallel()
	for _, internal := range []bool{false, true} {
		doTestSitemapOutput(t, internal)
	}
}

func doTestSitemapOutput(t *testing.T, internal bool) {

	cfg, fs := newTestCfg()
	cfg.Set("baseURL", "http://auth/bub/")

	depsCfg := deps.DepsCfg{Fs: fs, Cfg: cfg}

	depsCfg.WithTemplate = func(templ tpl.TemplateHandler) error {
		if !internal {
			templ.AddTemplate("sitemap.xml", sitemapTemplate)
		}

		// We want to check that the 404 page is not included in the sitemap
		// output. This template should have no effect either way, but include
		// it for the clarity.
		templ.AddTemplate("404.html", "Not found")
		return nil
	}

	writeSourcesToSource(t, "content", fs, weightedSources...)
	s := buildSingleSite(t, depsCfg, BuildCfg{})
	th := testHelper{s.Cfg, s.Fs, t}
	outputSitemap := "public/sitemap.xml"

	th.assertFileContent(outputSitemap,
		// Regular page
		" <loc>http://auth/bub/sect/doc1/</loc>",
		// Home page
		"<loc>http://auth/bub/</loc>",
		// Section
		"<loc>http://auth/bub/sect/</loc>",
		// Tax terms
		"<loc>http://auth/bub/categories/</loc>",
		// Tax list
		"<loc>http://auth/bub/categories/hugo/</loc>",
	)

	content := readDestination(th.T, th.Fs, outputSitemap)
	require.NotContains(t, content, "404")

}

func TestParseSitemap(t *testing.T) {
	t.Parallel()
	expected := Sitemap{Priority: 3.0, Filename: "doo.xml", ChangeFreq: "3"}
	input := map[string]interface{}{
		"changefreq": "3",
		"priority":   3.0,
		"filename":   "doo.xml",
		"unknown":    "ignore",
	}
	result := parseSitemap(input)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Got \n%v expected \n%v", result, expected)
	}

}
