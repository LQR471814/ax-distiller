## ax-distiller

> distill text and various semantic information from websites, cutting out noise and styling with headless chrome and the accessibility tree.
> essentially a glorified website -> markdown convertor for now.

### usage

```bash
# setup headless chrome and ublock origin
go run cmd/setup/main.go

# run the distiller
cd cmd/distill-test
go build && ./distill-test

# check the results
cat out_en.wikipedia.org.md
```

