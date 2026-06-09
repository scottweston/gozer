# Gozer

Gozer is a fast & simple static site generator written in Golang.

- Converts Markdown and [djot](https://www.djot.net) to HTML.
- Allows you to use page-specific templates.
- Creates an XML sitemap for search engines.
- Creates an RSS feed for feed readers.

The [example](example/) directory contains a barebones example of a Gozer site.

This fork adds the following behavior beyond upstream `main`:

- Content files ending in `.md`, `.dj`, or `.html` are evaluated as Go templates before conversion or output.
- Arbitrary TOML `data` can be defined globally in `config.toml` and per page in front matter, then accessed as `.Site.Data` and `.Page.Data`.
- Content-template parse and execution errors include the source file path.
- Djot conversion errors are returned with the source file and, when possible, the line near the parser failure.
- `Prev` and `Next` template variables are populated for dated posts.
- `gozer -version` prints the `djot-content-templating` feature token for programmatic checks.

Sample websites using Gozer:

- [Simplest possible example](example/)
- My personal website: [site](https://www.dannyvankooten.com/) - [source](https://github.com/dannyvankooten/www.dannyvankooten.com)

## Installation

You can install Gozer by first installing a Go compiler and then running:

```sh
go install github.com/dannyvankooten/gozer@latest
```

## Usage

Run `gozer new` to quickly generate an empty directory structure.

```txt
├── config.toml                # Configuration file
├── content                    # Posts and pages
│   └── index.md
├── public                     # Static files
└── templates                  # Template files
    └── default.html
```

Then, run `gozer build` to generate your site.

Any supported content files placed in your `content/` directory will result in an HTML page in your build directory after running `gozer build`.

For example:

- `content/index.md` creates a file `build/index.html` so it is accessible over HTTP at `/`
- `content/about.md` creates a file `build/about/index.html` so it is accessible over HTTP at `/about/`.


## Commands

Run `gozer` without any arguments to view the help text.

```
Gozer - a fast & simple static site generator

Usage: gozer [OPTIONS] <COMMAND>

Commands:
    build   Deletes the output directory if there is one and builds the site
    serve   Builds the site and starts an HTTP server on http://localhost:8080
    watch   Builds the site and watches for file changes
    new     Creates a new site structure in the given directory

Options:
    -r, --root <ROOT> Directory to use as root of project (default: .)
    -c, --config <CONFIG> Path to configuration file (default: config.toml)
        --version Print version and build features
        --listen <INTERFACE:PORT> Interface to listen on; only used with 'serve',
                 'INTERFACE' is optional. e.g. '--listen :9000 serve'

```

Use `gozer -version` to print the binary version and build features. This fork prints `djot-content-templating`, which can be checked programmatically:

```sh
gozer -version | grep djot-content-templating
```

## Content files

Content files in your `content/` directory can end in `.md`, `.dj`, or `.html`. Markdown and Djot files are converted to HTML; HTML files are used directly after template evaluation.

Each content file can have TOML front matter specifying the page title:

```md
+++
title = "My page title"
+++

Page content here.
```

**djot note** djot has not settled on a syntax for front matter. Until [issue #35](https://github.com/jgm/djot/issues/35) is resolved, TOML front matter in djot documents are used.

When Djot conversion fails inside `godjot`, Gozer reports the file being processed and, when possible, the line near the parser failure.

### Templates
The default template for every page is `default.html`. You can override it by setting the `template` variable in your front matter.

```md
+++
title = "My page title"
template = "special-page.html"
+++

Page content here.
```

Templates are powered by Go's standard `html/template` package, so you can use all the [actions described here](https://pkg.go.dev/text/template#hdr-Actions).

Content files are also evaluated as templates before Markdown or Djot conversion, or before HTML content is inserted into the page template. This applies to generated pages and RSS feed descriptions. Content templates use the same variables and template functions as page templates. During content-file template evaluation, `Content` is available but empty to avoid recursive rendering. Content-template parse and execution errors include the source file path.

Every template receives the following set of variables:

```
Pages       # Slice of all pages in the site
Posts       # Slice of all posts in the site (any page with a date in the filename)
Site        # Global site properties: Url, Title, Data
Meta        # All keys from config.toml (for example: title, url, custom fields)
Page        # The current page: Title, Permalink, UrlPath, DatePublished, DateModified, Meta, Data
Prev        # Previous dated post, or nil
Next        # Next dated post, or nil
Title       # The current page title, shorthand for Page.Title
Content     # The current page's HTML content.
Now         # Timestamp of build, instance of time.Time
SiteUrl     # Deprecated alias for Site.Url
```

`Meta` contains all keys from your `config.toml`. Front matter keys are available on `Page.Meta`.

```gotemplate
<meta name="description" content="{{ index .Meta "description" }}">

{{ if index .Page.Meta "draft" }}
    <p>This page is still a draft.</p>
{{ end }}
```

Arbitrary data can be defined under the `data` namespace in `config.toml` and page front matter. Global data is available as `.Site.Data`, and page-specific data is available as `.Page.Data`.

```toml
[[data.foobar]]
name = "namey mcnamster"
url = "https://example.com/namey"

[[data.foobar]]
name = "boaty mcboatface"
url = "https://example.com/boaty"
```

```gotemplate
{{ range .Site.Data.foobar }}
    <a href="{{ .url }}">{{ .name }}</a>
{{ end }}
```

```md
+++
title = "My page title"

[[data.links]]
name = "Example"
url = "https://example.com"
+++
```

```gotemplate
{{ range .Page.Data.links }}
    <a href="{{ .url }}">{{ .name }}</a>
{{ end }}
```

The custom template functions available in page templates and content templates are:

```
HasPrefix   # strings.HasPrefix
HasSuffix   # strings.HasSuffix
Contains    # strings.Contains
Replace     # strings.Replace
GroupByDate # Groups pages by a Go time format, for example "2006" or "January"
```

The `Page` variable is an instance of the object below:

```
type Page struct {
    // Title of this page
    Title         string

    // Template this page uses for rendering. Defaults to "default.html".
    Template      string

    // Time this page was published (parsed from file name).
    DatePublished time.Time

    // Time this page was last modified on the filesystem.
    DateModified  time.Time

    // The full URL to this page, including the site URL.
    Permalink     string

    // URL path for this page, relative to site URL
    UrlPath       string

    // Path to source file for this page, relative to content root
    Filepath      string

    // Parsed front matter values, keyed by TOML key
    Meta          map[string]any

    // Deprecated: use Meta.
    Attrs         map[string]any

    // Arbitrary page data from front matter
    Data          map[string]any
}
```

To show a list of the 5 most recent posts:

```gotemplate
{{ range (slice .Posts 0 5) }}
    <a href="{{ .Permalink }}">{{ .Title }}</a> <small>{{ .DatePublished.Format "Jan 02, 2006" }}</small><br />
{{ end }}
```

## Contributing

Gozer development happens on [GitHub](https://github.com/).

- [Code repository](https://github.com/dannyvankooten/gozer)
- [Issue tracker](https://github.com/dannyvankooten/gozer/issues)

## License

Gozer is open-sourced under the MIT license.
