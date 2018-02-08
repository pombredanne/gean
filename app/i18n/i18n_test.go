// Copyright 2017 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package i18n

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/gostores/configurator"
	"github.com/gostores/fsintra"
	"github.com/gostores/notepad"
	"github.com/gostores/require"

	"github.com/geego/gean/app/config"
	"github.com/geego/gean/app/deps"
	"github.com/geego/gean/app/geanfs"
	"github.com/geego/gean/app/helpers"
	"github.com/geego/gean/app/tpl/tplimpl"
)

var logger = notepad.NewNotepad(notepad.LevelError, notepad.LevelError, os.Stdout, ioutil.Discard, "", log.Ldate|log.Ltime)

type i18nTest struct {
	data                             map[string][]byte
	args                             interface{}
	lang, id, expected, expectedFlag string
}

var i18nTests = []i18nTest{
	// All translations present
	{
		data: map[string][]byte{
			"en.toml": []byte("[hello]\nother = \"Hello, World!\""),
			"es.toml": []byte("[hello]\nother = \"¡Hola, Mundo!\""),
		},
		args:         nil,
		lang:         "es",
		id:           "hello",
		expected:     "¡Hola, Mundo!",
		expectedFlag: "¡Hola, Mundo!",
	},
	// Translation missing in current language but present in default
	{
		data: map[string][]byte{
			"en.toml": []byte("[hello]\nother = \"Hello, World!\""),
			"es.toml": []byte("[goodbye]\nother = \"¡Adiós, Mundo!\""),
		},
		args:         nil,
		lang:         "es",
		id:           "hello",
		expected:     "Hello, World!",
		expectedFlag: "[i18n] hello",
	},
	// Translation missing in default language but present in current
	{
		data: map[string][]byte{
			"en.toml": []byte("[goodbye]\nother = \"Goodbye, World!\""),
			"es.toml": []byte("[hello]\nother = \"¡Hola, Mundo!\""),
		},
		args:         nil,
		lang:         "es",
		id:           "hello",
		expected:     "¡Hola, Mundo!",
		expectedFlag: "¡Hola, Mundo!",
	},
	// Translation missing in both default and current language
	{
		data: map[string][]byte{
			"en.toml": []byte("[goodbye]\nother = \"Goodbye, World!\""),
			"es.toml": []byte("[goodbye]\nother = \"¡Adiós, Mundo!\""),
		},
		args:         nil,
		lang:         "es",
		id:           "hello",
		expected:     "",
		expectedFlag: "[i18n] hello",
	},
	// Default translation file missing or empty
	{
		data: map[string][]byte{
			"en.toml": []byte(""),
		},
		args:         nil,
		lang:         "es",
		id:           "hello",
		expected:     "",
		expectedFlag: "[i18n] hello",
	},
	// Context provided
	{
		data: map[string][]byte{
			"en.toml": []byte("[wordCount]\nother = \"Hello, {{.WordCount}} people!\""),
			"es.toml": []byte("[wordCount]\nother = \"¡Hola, {{.WordCount}} gente!\""),
		},
		args: struct {
			WordCount int
		}{
			50,
		},
		lang:         "es",
		id:           "wordCount",
		expected:     "¡Hola, 50 gente!",
		expectedFlag: "¡Hola, 50 gente!",
	},
	// Same id and translation in current language
	// https://github.com/geego/gean/app/issues/2607
	{
		data: map[string][]byte{
			"es.toml": []byte("[hello]\nother = \"hello\""),
			"en.toml": []byte("[hello]\nother = \"hi\""),
		},
		args:         nil,
		lang:         "es",
		id:           "hello",
		expected:     "hello",
		expectedFlag: "hello",
	},
	// Translation missing in current language, but same id and translation in default
	{
		data: map[string][]byte{
			"es.toml": []byte("[bye]\nother = \"bye\""),
			"en.toml": []byte("[hello]\nother = \"hello\""),
		},
		args:         nil,
		lang:         "es",
		id:           "hello",
		expected:     "hello",
		expectedFlag: "[i18n] hello",
	},
	// Unknown language code should get its plural spec from en
	{
		data: map[string][]byte{
			"en.toml": []byte(`[readingTime]
one ="one minute read"
other = "{{.Count}} minutes read"`),
			"klingon.toml": []byte(`[readingTime]
one =  "eitt minutt med lesing"
other = "{{ .Count }} minuttar lesing"`),
		},
		args:         3,
		lang:         "klingon",
		id:           "readingTime",
		expected:     "3 minuttar lesing",
		expectedFlag: "3 minuttar lesing",
	},
}

func doTestI18nTranslate(t *testing.T, test i18nTest, cfg config.Provider) string {
	assert := require.New(t)
	fs := geanfs.NewMem(cfg)
	tp := NewTranslationProvider()
	depsCfg := newDepsConfig(tp, cfg, fs)
	d, err := deps.New(depsCfg)
	assert.NoError(err)

	for file, content := range test.data {
		err := fsintra.WriteFile(fs.Source, filepath.Join("i18n", file), []byte(content), 0755)
		assert.NoError(err)
	}

	assert.NoError(d.LoadResources())
	f := tp.t.Func(test.lang)
	return f(test.id, test.args)

}

func newDepsConfig(tp *TranslationProvider, cfg config.Provider, fs *geanfs.Fs) deps.DepsCfg {
	l := helpers.NewLanguage("en", cfg)
	l.Set("i18nDir", "i18n")
	return deps.DepsCfg{
		Language:            l,
		Cfg:                 cfg,
		Fs:                  fs,
		Logger:              logger,
		TemplateProvider:    tplimpl.DefaultTemplateProvider,
		TranslationProvider: tp,
	}
}

func TestI18nTranslate(t *testing.T) {
	var actual, expected string
	v := configurator.New()
	v.SetDefault("defaultContentLanguage", "en")

	// Test without and with placeholders
	for _, enablePlaceholders := range []bool{false, true} {
		v.Set("enableMissingTranslationPlaceholders", enablePlaceholders)

		for _, test := range i18nTests {
			if enablePlaceholders {
				expected = test.expectedFlag
			} else {
				expected = test.expected
			}
			actual = doTestI18nTranslate(t, test, v)
			require.Equal(t, expected, actual)
		}
	}
}
