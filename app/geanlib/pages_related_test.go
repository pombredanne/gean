package geanlib

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gostores/require"
	"yiqilai.tech/gean/app/common/types"
	"yiqilai.tech/gean/app/deps"
)

func TestRelated(t *testing.T) {
	assert := require.New(t)

	t.Parallel()

	var (
		cfg, fs = newTestCfg()
		//th      = testHelper{cfg, fs, t}
	)

	pageTmpl := `---
title: Page %d
keywords: [%s]
date: %s
---

Content
`

	writeSource(t, fs, filepath.Join("content", "page1.md"), fmt.Sprintf(pageTmpl, 1, "hugo, says", "2017-01-03"))
	writeSource(t, fs, filepath.Join("content", "page2.md"), fmt.Sprintf(pageTmpl, 2, "hugo, rocks", "2017-01-02"))
	writeSource(t, fs, filepath.Join("content", "page3.md"), fmt.Sprintf(pageTmpl, 3, "bep, says", "2017-01-01"))

	s := buildSingleSite(t, deps.DepsCfg{Fs: fs, Cfg: cfg}, BuildCfg{SkipRender: true})
	assert.Len(s.RegularPages, 3)

	result, err := s.RegularPages.RelatedTo(types.NewKeyValuesStrings("keywords", "hugo", "rocks"))

	assert.NoError(err)
	assert.Len(result, 2)
	assert.Equal("Page 2", result[0].Title)
	assert.Equal("Page 1", result[1].Title)

	result, err = s.RegularPages.Related(s.RegularPages[0])
	assert.Len(result, 2)
	assert.Equal("Page 2", result[0].Title)
	assert.Equal("Page 3", result[1].Title)

	result, err = s.RegularPages.RelatedIndices(s.RegularPages[0], "keywords")
	assert.Len(result, 2)
	assert.Equal("Page 2", result[0].Title)
	assert.Equal("Page 3", result[1].Title)

	result, err = s.RegularPages.RelatedTo(types.NewKeyValuesStrings("keywords", "bep", "rocks"))
	assert.NoError(err)
	assert.Len(result, 2)
	assert.Equal("Page 2", result[0].Title)
	assert.Equal("Page 3", result[1].Title)
}