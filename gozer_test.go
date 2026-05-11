package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"
	"testing"
	"time"
)

func TestExampleSite(t *testing.T) {
	_ = os.RemoveAll("build/")
	buildSite("example/", "config.toml")

	tests := []struct {
		file     string
		contains [][]byte
	}{
		{"index.html", [][]byte{
			[]byte("<p>Hey, welcome on my site!</p>"),
			[]byte("<title>My site</title>"),
			[]byte("<li>key1: 1</li>"),
			[]byte("<li>key2: two</li>"),
			[]byte("<li>title: My site</li>")},
		},
		{"about/index.html", [][]byte{
			[]byte("<title>About me</title>"),
			[]byte("<li>draft: true</li>"),
			[]byte("<li>tags: [about gozer]</li>"),
			[]byte("<li>Dolor</li>")},
		},
		{"hello-world/index.html", [][]byte{
			[]byte("<title>Hello, world!</title>"),
			[]byte("This is a blog post.")},
		},
		{"favicon.ico", nil},
		{"feed.xml", [][]byte{
			[]byte("<item><title>Hello, world!</title><link>http://localhost:8080/hello-world/</link>"),
		}},
		{"sitemap.xml", [][]byte{
			[]byte("<url><loc>http://localhost:8080/</loc>"),
		}},
		{"sitemap.xsl", nil},
	}

	for _, tc := range tests {
		content, err := os.ReadFile("build/" + tc.file)
		if err != nil {
			t.Errorf("Expected file, got error: %s", err)
		}

		for _, e := range tc.contains {
			if !bytes.Contains(content, e) {
				t.Errorf("Output file %s does not have expected content %s", tc.file, e)
			}
		}

	}
}

func TestParseConfigFile(t *testing.T) {
	s := Site{}
	if err := parseConfig(&s, "example/config.toml"); err != nil {
		t.Errorf("error parsing config file: %s", err)
	}

	expectedSiteUrl := "http://localhost:8080/"
	expectedTitle := "My website"
	if s.SiteUrl != expectedSiteUrl {
		t.Errorf("invalid site url. expected %v, got %v", expectedSiteUrl, s.SiteUrl)
	}

	if s.Title != expectedTitle {
		t.Errorf("invalid site title. expected %v, got %v", expectedTitle, s.Title)
	}

	if s.Data != nil {
		t.Errorf("expected optional config data to be nil, got %v", s.Data)
	}
}

func TestParseConfigFileData(t *testing.T) {
	file := t.TempDir() + "/config.toml"
	err := os.WriteFile(file, []byte(`url = "https://example.com"
title = "My website"

[[data.foobar]]
name = "namey mcnamster"
url = "https://example.com/namey"

[[data.foobar]]
name = "boaty mcboatface"
url = "https://example.com/boaty"
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	s := Site{}
	if err := parseConfig(&s, file); err != nil {
		t.Errorf("error parsing config file: %s", err)
	}

	links := dataList(t, s.Data, "foobar")
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
	if links[0]["name"] != "namey mcnamster" || links[0]["url"] != "https://example.com/namey" {
		t.Errorf("invalid first link: %v", links[0])
	}
	if links[1]["name"] != "boaty mcboatface" || links[1]["url"] != "https://example.com/boaty" {
		t.Errorf("invalid second link: %v", links[1])
	}
}

func TestParseConfigMetaAndAttrsAlias(t *testing.T) {
	s := Site{}
	if err := parseConfig(&s, "example/config.toml"); err != nil {
		t.Fatal(err)
	}

	if got := s.Meta["key1"]; fmt.Sprint(got) != "1" {
		t.Errorf("invalid site attr key1. expected %v, got %v", 1, got)
	}

	if got := s.Meta["key2"]; got != "two" {
		t.Errorf("invalid site attr key2. expected %q, got %v", "two", got)
	}

	if got := s.Meta["title"]; got != "My website" {
		t.Errorf("invalid site attr title. expected %q, got %v", "My website", got)
	}

	if got := s.Meta["url"]; got != "http://localhost:8080" {
		t.Errorf("invalid site attr url. expected %q, got %v", "http://localhost:8080", got)
	}

	if got := s.Attrs["key1"]; fmt.Sprint(got) != "1" {
		t.Errorf("invalid site alias attr key1. expected %v, got %v", 1, got)
	}

	s.Meta["alias"] = "yes"
	if got := s.Attrs["alias"]; got != "yes" {
		t.Errorf("expected attrs alias to reflect meta writes, got %v", got)
	}

	s.Attrs["alias2"] = "yes"
	if got := s.Meta["alias2"]; got != "yes" {
		t.Errorf("expected meta to reflect attrs alias writes, got %v", got)
	}
}

func TestParseFrontMatter(t *testing.T) {
	p := &Page{
		Filepath: "example/content/index.md",
	}

	if err := parseFrontMatter(p); err != nil {
		t.Fatal(err)
	}

	expectedTitle := "My site"
	if p.Title != expectedTitle {
		t.Errorf("Invalid title. Expected %v, got %v", expectedTitle, p.Title)
	}

	if got := p.Meta["title"]; got != expectedTitle {
		t.Errorf("Invalid front matter title attr. Expected %v, got %v", expectedTitle, got)
	}

	if got := p.Attrs["title"]; got != expectedTitle {
		t.Errorf("Invalid front matter title alias attr. Expected %v, got %v", expectedTitle, got)
	}

	if p.Data != nil {
		t.Errorf("expected optional page data to be nil, got %v", p.Data)
	}
}

func TestParseFrontMatterData(t *testing.T) {
	file := t.TempDir() + "/page.md"
	err := os.WriteFile(file, []byte(`+++
title = "My data page"

[[data.links]]
name = "Example"
url = "https://example.com"
+++

Page content here.
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	p := &Page{
		Filepath: file,
	}
	if err := parseFrontMatter(p); err != nil {
		t.Fatal(err)
	}

	links := dataList(t, p.Data, "links")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0]["name"] != "Example" || links[0]["url"] != "https://example.com" {
		t.Errorf("invalid link: %v", links[0])
	}
}

func TestParseFrontMatterDjot(t *testing.T) {
	p := &Page{
		Filepath: "example/content/djot.dj",
	}

	if err := parseFrontMatter(p); err != nil {
		t.Fatal(err)
	}

	expectedTitle := "Multiple markup support"
	if p.Title != expectedTitle {
		t.Errorf("Invalid title. Expected %v, got %v", expectedTitle, p.Title)
	}
}

func TestParseFrontMatterMetaAndAttrsAlias(t *testing.T) {
	p := &Page{
		Filepath: "example/content/about.md",
	}

	if err := parseFrontMatter(p); err != nil {
		t.Fatal(err)
	}

	if got := p.Meta["draft"]; got != true {
		t.Errorf("Invalid draft attr. Expected true, got %v", got)
	}

	if got := p.Meta["title"]; got != "About me" {
		t.Errorf("Invalid title attr. Expected %q, got %v", "About me", got)
	}

	if got := fmt.Sprint(p.Meta["tags"]); got != "[about gozer]" {
		t.Errorf("Invalid tags attr. Expected %q, got %v", "[about gozer]", got)
	}

	p.Meta["alias"] = "yes"
	if got := p.Attrs["alias"]; got != "yes" {
		t.Errorf("expected attrs alias to reflect meta writes, got %v", got)
	}

	p.Attrs["alias2"] = "yes"
	if got := p.Meta["alias2"]; got != "yes" {
		t.Errorf("expected meta to reflect attrs alias writes, got %v", got)
	}
}

func TestParseContent(t *testing.T) {
	p := &Page{
		Filepath: "example/content/index.md",
	}

	content, err := p.ParseContent()
	if err != nil {
		t.Fatal(err)
	}

	if content != "<p>Hey, welcome on my site!</p>\n" {
		t.Errorf("Invalid content. Got %v", content)
	}
}

func TestParseContentDjot(t *testing.T) {
	p := &Page{
		Filepath: "example/content/djot_test.dj",
	}

	content, err := p.ParseContent()
	if err != nil {
		t.Fatal(err)
	}

	if content != "<p>Hey, welcome on my site!</p>\n" {
		t.Errorf("Invalid content. Got %v", content)
	}
}

func TestConvertDjot(t *testing.T) {
	content, err := ConvertDjot([]byte("Hey, welcome on my site!"))
	if err != nil {
		t.Fatal(err)
	}

	if content != "<p>Hey, welcome on my site!</p>\n" {
		t.Errorf("Invalid content. Got %v", content)
	}
}

func TestTemplateData(t *testing.T) {
	oldTemplates := templates
	defer func() {
		templates = oldTemplates
		_ = os.RemoveAll("build/data-test")
	}()

	var err error
	templates, err = template.New("gozer").Parse(`{{ define "default.html" }}{{ range .Site.Data.foobar }}{{ .name }}={{ .url }};{{ end }}|{{ range .Page.Data.links }}{{ .name }}={{ .url }};{{ end }}{{ end }}`)
	if err != nil {
		t.Fatal(err)
	}

	file := t.TempDir() + "/page.html"
	if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	s := Site{
		SiteUrl: "https://example.com/",
		Title:   "My website",
		Data: map[string]any{
			"foobar": []map[string]any{{
				"name": "global",
				"url":  "https://example.com/global",
			}},
		},
	}
	p := &Page{
		Filepath:  file,
		UrlPath:   "data-test/",
		Template:  "default.html",
		Permalink: "https://example.com/data-test/",
		Data: map[string]any{
			"links": []map[string]any{{
				"name": "page",
				"url":  "https://example.com/page",
			}},
		},
	}

	if err := s.buildPage(p); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile("build/data-test/index.html")
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte("global=https://example.com/global;|page=https://example.com/page;")
	if !bytes.Contains(content, expected) {
		t.Errorf("expected template data output %q, got %q", expected, content)
	}
}

func TestContentTemplateDataBeforeMarkdownConversion(t *testing.T) {
	oldNow := now
	now = time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	defer func() {
		now = oldNow
	}()

	file := t.TempDir() + "/page.md"
	if err := os.WriteFile(file, []byte(`+++
title = "Content template page"
+++

Site: {{ .Site.Url }} {{ .Site.Title }} {{ range .Site.Data.links }}{{ .name }}{{ end }}
Page: {{ .Page.Title }} {{ .Page.Permalink }} {{ range .Page.Data.links }}{{ .name }}{{ end }}
Counts: {{ len .Posts }} {{ len .Pages }}
Now: {{ .Now.Year }}
Legacy: {{ .SiteUrl }}
{{ "**rendered markdown**" }}
`), 0644); err != nil {
		t.Fatal(err)
	}

	p := Page{
		Filepath:  file,
		Title:     "Content template page",
		Permalink: "https://example.com/page/",
		Data: map[string]any{
			"links": []map[string]any{{"name": "page-link"}},
		},
	}
	s := Site{
		SiteUrl: "https://example.com/",
		Title:   "Example site",
		Pages:   []Page{p, {Title: "Other page"}},
		Posts:   []Page{p},
		Data: map[string]any{
			"links": []map[string]any{{"name": "site-link"}},
		},
	}

	content, err := s.parsePageContent(&p, template.HTML(""))
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		`Site: <a href="https://example.com/">https://example.com/</a> Example site site-link`,
		`Page: Content template page <a href="https://example.com/page/">https://example.com/page/</a> page-link`,
		"Counts: 1 2",
		"Now: 2026",
		`Legacy: <a href="https://example.com/">https://example.com/</a>`,
		"<strong>rendered markdown</strong>",
	}
	for _, e := range expected {
		if !strings.Contains(content, e) {
			t.Errorf("expected content template output %q in %q", e, content)
		}
	}
}

func TestContentTemplateContentVariableIsEmpty(t *testing.T) {
	file := t.TempDir() + "/page.md"
	if err := os.WriteFile(file, []byte(`{{ if .Content }}has content{{ else }}empty content{{ end }}`), 0644); err != nil {
		t.Fatal(err)
	}

	p := Page{Filepath: file}
	content, err := (&Site{}).parsePageContent(&p, template.HTML(""))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(content, "<p>empty content</p>") {
		t.Errorf("expected empty content marker, got %q", content)
	}
}

func TestContentTemplateHTML(t *testing.T) {
	file := t.TempDir() + "/page.html"
	if err := os.WriteFile(file, []byte(`<h1>{{ .Title }}</h1>`), 0644); err != nil {
		t.Fatal(err)
	}

	p := Page{
		Filepath: file,
		Title:    "HTML template page",
	}
	content, err := (&Site{}).parsePageContent(&p, template.HTML(""))
	if err != nil {
		t.Fatal(err)
	}

	if content != "<h1>HTML template page</h1>" {
		t.Errorf("expected rendered HTML content, got %q", content)
	}
}

func TestContentTemplateSyntaxError(t *testing.T) {
	file := t.TempDir() + "/page.md"
	if err := os.WriteFile(file, []byte(`{{ if .Title }}`), 0644); err != nil {
		t.Fatal(err)
	}

	p := Page{Filepath: file}
	_, err := (&Site{}).parsePageContent(&p, template.HTML(""))
	if err == nil {
		t.Fatal("expected content template syntax error")
	}
	if !strings.Contains(err.Error(), "parsing content template "+file) {
		t.Errorf("expected filepath in content template error, got %q", err)
	}
}

func TestFirstFailingLine(t *testing.T) {
	content := []byte("line 1\nline 2\nline 3\n")
	line := firstFailingLine(content, func(prefix []byte) bool {
		return strings.Contains(string(prefix), "line 2")
	})

	if line != 2 {
		t.Errorf("expected line 2, got %d", line)
	}
}

func TestFirstFailingLineNoMatch(t *testing.T) {
	line := firstFailingLine([]byte("line 1\nline 2\n"), func(prefix []byte) bool {
		return false
	})

	if line != 0 {
		t.Errorf("expected line 0, got %d", line)
	}
}

func TestFilepathToUrlpath(t *testing.T) {
	tests := []struct {
		input                 string
		expectedUrlPath       string
		expectedDatePublished time.Time
	}{
		{input: "content/index.md", expectedUrlPath: "", expectedDatePublished: time.Time{}},
		{input: "content/about.md", expectedUrlPath: "about/", expectedDatePublished: time.Time{}},
		{input: "content/blog/index.md", expectedUrlPath: "blog/", expectedDatePublished: time.Time{}},
		{input: "content/projects/gozer.md", expectedUrlPath: "projects/gozer/", expectedDatePublished: time.Time{}},
		{input: "content/2023-11-23-hello-world.md", expectedUrlPath: "hello-world/", expectedDatePublished: time.Date(2023, 11, 23, 0, 0, 0, 0, time.UTC)},
		{input: "content/blog/2023-11-23-here-we-are.md", expectedUrlPath: "blog/here-we-are/", expectedDatePublished: time.Date(2023, 11, 23, 0, 0, 0, 0, time.UTC)},
		{input: "content/blog/2023-11-23-there-we-were.dj", expectedUrlPath: "blog/there-we-were/", expectedDatePublished: time.Date(2023, 11, 23, 0, 0, 0, 0, time.UTC)},
		{input: "content/2023-11-23-hello-sunshine.dj", expectedUrlPath: "hello-sunshine/", expectedDatePublished: time.Date(2023, 11, 23, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range tests {
		urlPath, datePublished := parseFilename(tc.input, "")
		if urlPath != tc.expectedUrlPath {
			t.Errorf("expected %v, got %v", tc.expectedUrlPath, urlPath)
		}

		if !datePublished.Equal(tc.expectedDatePublished) {
			t.Errorf("expected %v, got %v", tc.expectedDatePublished, datePublished)
		}
	}
}

func BenchmarkParseFrontMatter(b *testing.B) {
	data := `+++
title = "My page title"
template = "Page template"
+++

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur ac pretium magna. Phasellus ut ligula vel erat dictum sollicitudin eu a dolor. Donec orci mauris, cursus eget elementum eu, tempor sed massa. Aliquam mattis ullamcorper metus, sodales fermentum lectus fringilla id. Duis dui ligula, lobortis ut leo id, semper ultricies justo. Etiam vehicula sit amet ligula vitae maximus. Aenean consectetur nisl ac est convallis, vel dictum nulla iaculis. Ut dignissim lobortis ipsum, vel molestie lectus ornare quis. In hac habitasse platea dictumst. Sed ut elementum nulla.

Ut eleifend felis lacus, id condimentum purus laoreet et. Nam sodales mi cursus, porta enim aliquet, venenatis quam. Sed et quam nisl. Donec libero ex, eleifend sit amet dui at, fermentum semper sem. Donec gravida id nibh eu mollis. Fusce pellentesque gravida ipsum, sit amet sagittis tellus. Donec consectetur nulla enim. Donec quis ornare tellus. Maecenas eget imperdiet lacus. Ut imperdiet dui nisi, a tristique metus sodales vel.

Proin non ex id erat feugiat imperdiet. Duis posuere finibus quam, quis blandit lorem vehicula sit amet. Fusce pulvinar commodo magna, ut sodales massa interdum at. Nam in dapibus nunc. Vestibulum nisi nisl, vestibulum ac vestibulum in, maximus vitae nibh. Nulla egestas pellentesque velit, vitae tempor massa scelerisque et. Nam nibh metus, vestibulum eu justo vitae, venenatis faucibus eros. Suspendisse sed ligula dolor. Etiam sed elit ullamcorper, placerat ex at, rutrum est. Quisque vitae dolor non metus lobortis sagittis. Donec pretium orci aliquam tortor blandit, sed consectetur massa ullamcorper. Praesent vehicula nunc quis urna tempor finibus. Etiam lectus urna, tempor eget diam ac, consequat commodo nunc. Nullam sed venenatis ante, ac mollis lorem. Nunc vitae faucibus lectus. Vivamus vel arcu justo.

Quisque rhoncus elementum sapien ac semper. Sed tristique elit vel nibh semper tincidunt. Nunc feugiat massa eget magna accumsan, eu commodo tortor accumsan. Morbi porttitor metus nec tellus bibendum, in consectetur ex rhoncus. Phasellus dapibus tincidunt ligula, in posuere ligula. Praesent vestibulum porttitor lorem nec mollis. Praesent magna dui, bibendum sed malesuada a, tristique id ipsum. Phasellus non ipsum eget est vulputate rutrum non nec leo. Nam consequat lobortis lorem non accumsan. Aliquam eget elit in dolor malesuada consequat. Sed luctus bibendum arcu eu posuere. Integer nec ipsum turpis. Sed luctus risus ante, eget gravida magna auctor eget sodales sed.
`

	filepath := os.TempDir() + "/front-matter.md"
	err := os.WriteFile(filepath, []byte(data), 0655)
	if err != nil {
		return
	}

	p := &Page{
		Filepath: filepath,
	}
	for n := 0; n < b.N; n++ {
		if err := parseFrontMatter(p); err != nil {
			b.Error(err)
		}
	}
}

func dataList(t *testing.T, data map[string]any, key string) []map[string]any {
	t.Helper()

	raw, ok := data[key]
	if !ok {
		t.Fatalf("expected data key %q in %v", key, data)
	}

	links, ok := raw.([]map[string]any)
	if !ok {
		t.Fatalf("expected data key %q to be []map[string]any, got %T", key, raw)
	}

	return links
}
