# ax-distiller

> efficiently convert websites into human and LLM friendly formats with the web accessibility tree.

## demos

### with AI markdown conversion

> This is a demo of converting a somewhat involved website into an human/LLM-friendly markdown file by feeding a dump of the accessibility tree to an LLM. This allows you to retain layout and semantic context in the output file.

- [markdown result](./examples/drawabox_(with_gemini_postprocessing).md)
- [accessibility tree dump](./examples/drawabox_axtree_dump.xml)

For reference, this is prompt used:

```md
# Instructions

- Render the following accessibility tree of a website in markdown.
- Take liberties with layout and structuring.
- Make sure images are linked to, but do not render them with the "!" qualifier.

## Chrome Accessibility Tree

{XML representation of AX tree}
```

### with heuristic markdown conversion

> This is a demo of converting a wikipedia article into a human/LLM-friendly markdown file through the use of a variety of heuristics. This is much faster than feeding an accessibility tree to an LLM but does not retain semantic or layout information.

- [markdown result](./examples/wikipedia_sample.md)

## usage

```bash
# setup headless chrome and ublock origin
go run ./cmd/setup

# dump the accessibility tree of some website (with filtering of useless ax nodes)
go run ./cmd/dump-ax-tree "https://somewebsite.com/..." > serialized.xml

# run the distiller
cd cmd/distill-test
go build && ./distill-test

# check the results
cat out_en.wikipedia.org.md
```

