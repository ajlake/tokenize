# tokenize

A CLI I hacked together for use with FoundryVTT. Consumes JPEGs or PNGs and outputs icons in PNG format.
Nothing wrong with https://rolladvantage.com/tokenstamp/, but I wanted something command-line based,
completely offline, and more tailored to my workflow.

Usage:
```
tokenize path/to/image.jpg [images ...]
```

**Input**

```
example.jpg
```
![Example](/example/example.jpg)

```
tokenize example.jpg
```

**Output**

```
example_gold.png
example_silver.png
```
![Example](/example/example_gold.png)
![Example](/example/example_silver.png)
