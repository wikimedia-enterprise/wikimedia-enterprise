# Wikimedia Enterprise Parser

Wikimedia Enterprise Parser for Wikipedia HTML API

## Working with the repository

### Setting up the repo

We are using specific workflow for this repository cuz we always include it as a submodule and never use it as a standalone repo. That's why we have additional steps to make things work. List of steps:

1. Initialize the `go.mod` by running this command:

   ```bash
   go mod init wikimedia-enterprise/general/parser
   ```

1. Then initialize the packages by running:

   ```bash
   go mod tidy
   ```

### Unit testing

We are using snapshots library to generate uni tests. If you need to re-generate the snapshots just run. For some reason the snapshot file gets appended to each time we run `UPDATE_SNAPS=true`, this is causing git timeouts in our Gitlab CICD pipeline. We recommend using a ``rm -rf` to delete the old snaps and keep the file size manageable.

```bash
rm -rf .snapshots;UPDATE_SNAPSHOTS=true go test ./...
```

## Additional info

### How do we get the abstract currently?

First we get the first section of the body using [goquery](https://github.com/puerkitobio/goquery#examples)

###### Heuristics:

1. Remove generic unnecessary nodes from selection
2. Remove specific nodes with certain attributes/values
3. To safely process parenthesis, "flatten" the selection (DOM) - Replace certain elements with their text; remove some attributes; escape parenthesis from attributes
4. Retrieve Html out of the selection.
5. For parenthesis processing, check if there are any parenthesis that contain math expression or chemical formula.
6. If no math/chem parenthesis are found:
   (a) remove all nested parenthesis;
   (b)remove all the parenthesis that have any space inside them;
   ( c) remove all parenthesis that have non-Latin stuff inside.
7. Remove empty parenthesis/brackets. Fix spacing around punctuations.

###### Known leads issues:

| Article                     | Issue                                              | Issue seen with pageSummary API as well | Minor issue |
| --------------------------- | -------------------------------------------------- | --------------------------------------- | ----------- |
| Kyoto (enwiki)              | ommission of text in bold                          | No                                      | No          |
| David Cameron (frwiki)      | inclusion of pronunciation text in / /             | Yes                                     | -           |
| राजस्थान (hiwiki)           | inclusion of citation needed [कृपया उद्धरण जोड़ें] | Yes                                     | No          |
| S.L.\_Benfica (enwiki)      | space before ,                                     | No                                      | Yes         |
| Switzerland (enwiki)        | inclusion of info text `audio (help·info) `        | after 1st paragraph                     | No          |
| Niedersachsen (dewiki)      | extra space after removing the audio icon          | No                                      | Yes         |
| Scythian_languages (enwiki) | Excludes blockquote at the end of the section      | after 1st paragraphn                    | -           |
