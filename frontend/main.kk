module main

import std/os/path
import std/os/file
import html/html/html

infixl 1 (.=)

fun attr/(.=)(k : string, v : string)
    add-attr(k, v)

fun add-id(id-str)
    add-attr("id", id-str)

fun meta-name(name-str, content-str)
    with component("meta")
    add-attr("content", content-str)
    add-attr("name", name-str)

fun meta-property(property-str, content-str)
    with component("meta")
    add-attr("content", content-str)
    add-attr("property", property-str)

fun main-page () : _ component
    with html
    "lang" .= "en"

    head
      meta-charset  ()
      meta-viewport ()

      meta-name("description", "The frontend for my 100 backends project")
      meta-name("author",      "LitFill")
      meta-name("theme-color", "#5E3399")

      meta-property("og:title",       "Front end for 100 backends")
      meta-property("og:description", "The frontend for my 100 backends project")
      meta-property("og:type",        "website")

      title
        "Front end for 100 backends".text

      stylesheet("https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.violet.min.css")

      script-src("https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js", defer=True)

      link("icon", "/static/favicon.ico")
          add-attr("type", "image/x-icon")
          add-attr("sizes", "48x48")
      link("icon", "/static/favicon.svg")
          add-attr("type", "image/svg+xml")
          add-attr("sizes", "any")
      link("icon", "/static/favicon-96x96.png")
          add-attr("type", "image/png")
          add-attr("sizes", "96x96")
      link("apple-touch-icon", "/static/apple-touch-icon.png")
          add-attr("sizes", "180x180")
      link("manifest", "/site.webmanifest")
          add-attr("crossorigin", "use-credentials")

    with body
    header
      "class" .= "container"
      h1
        "URLs in server".text

    main
      "class"  .= "container"
      "role"   .= "main"
      "x-data" .= "urlService()"
      "x-init" .= "fetchURLs()"

      article
        form("http://localhost:8080/shorten", "POST")
          "x-data" .= "{ submitting: false }"
          "@submit" .= "submitting = true"

          fieldset
            "role" .= "group"

            input("original-url", "url", "")
              "placeholder" .= "https://example.com/very/long/url"
              set-bool-attr("required", True)
            button
              "type" .= "submit"
              ":disabled" .= "submitting"
              "Shorten" .text

      article
        input("search", "text", "")
          "x-model"     .= "search"
          "placeholder" .= "Filter by short or long URL..."
          ":disabled"   .= "items.length === 0"

        p
          "x-text" .= "statusMessage"
          "x-show" .= "statusMessage !== ''"
          "placeholder" .text
      p
        "x-show" .= "items.length > 0 && filteredItems.length === 0"
        "No URLs match your search." .text

      // TODO: a form for submitting a new link to /shorten

      div(classes=["overflow-auto"])
        with table
        "x-show" .= "filteredItems.length > 0"

        thead
          tr
            th
              "Short" .text
            th
              "Long"  .text

        tbody
          with template
          "x-for" .= "urlMap in filteredItems"
          ":key"  .= "urlMap.short"

          with tr
          td
            "x-text" .= "urlMap.short"
            "placeholder" .text
          td
            with "".a
            ":href"  .= "urlMap.long"
            "target" .= "_blank"
            "x-text" .= "urlMap.long"
            "placeholder" .text
    init-script ()

fun init-script()
    val script =
      r#"function urlService() {
        return {
          items:         [],
          loading:       false,
          search:        "",
          statusMessage: "Initializing...",

          get filteredItems() {
            if (!this.search.trim()) return this.items;
            const query = this.search.toLowerCase();
            return this.items.filter(item =>
              item.short.toLowerCase().includes(query) ||
              item.long.toLowerCase().includes(query)
            );
          },

          async fetchURLs() {
            this.loading       = true;
            this.statusMessage = "Loading items...";
            this.items         = [];
            try {
              const response = await fetch("http://localhost:8080/");
              if (!response.ok)
                throw new Error(`Server error: ${response.status}`);
              const data = await response.json();
              if (!data.length) this.statusMessage = "No items found.";
              else {
                this.items = data;
                this.statusMessage = "";
              }
            } catch (error) {
              this.statusMessage = `Error loading items: ${error.message}`;
            } finally {
              this.loading = false;
            }
          },
        };
      }"#
    text/script(script)

fun add-html-header (html-str)
    val header-str =
      "<!-- Generated by main.kk. Do not modify. -->\n<!DOCTYPE html>\n"
    header-str ++ html-str

fun main ()
    "=== Generating HTML ===".println
    val html-str = main-page().show.add-html-header
    html-str.println
    "index.html".path.write-text-file(html-str)
