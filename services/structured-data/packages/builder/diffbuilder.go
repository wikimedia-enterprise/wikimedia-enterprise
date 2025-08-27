package builder

import (
	"strings"
	"wikimedia-enterprise/services/structured-data/submodules/schema"
)

// getLongestRepeatedCharacter returns the length of the largest
// subsequently repeated characters substring from input string.
func getLongestRepeatedCharacter(text string) int {
	var lastChar rune
	var lastCharCount int
	var maxLastCharCount int
	text = strings.ToLower(text)

	for _, r := range text {
		if r == lastChar {
			lastCharCount++
			if lastCharCount > maxLastCharCount {
				maxLastCharCount = lastCharCount
			}
			continue
		}
		lastChar = r
		lastCharCount = 1
	}

	return maxLastCharCount
}

// DiffBuilder follows the builder pattern for the Diff schema.
type DiffBuilder struct {
	diff *schema.Diff
}

// NewDiffBuilder returns an diff builder with an empty diff.
func NewDiffBuilder() *DiffBuilder {
	return &DiffBuilder{
		diff: new(schema.Diff),
	}
}

// Size adds the diff size (in bytes) to the Diff schema.
func (db *DiffBuilder) Size(currWiki string, prevWiki string) *DiffBuilder {
	db.diff.Size = &schema.Size{
		UnitText: "B",
		Value:    float64(len([]byte(currWiki))) - float64(len([]byte(prevWiki))),
	}

	return db
}

// DictNonDictWords adds DictionaryWords and NonDictionaryWords to the Diff schema.
func (db *DiffBuilder) DictNonDictWords(parentDictWords, parentNonDictWords, currentDictWords, currentNonDictWords []string) *DiffBuilder {
	db.diff.DictionaryWords = NewDeltaBuilder().
		Current(currentDictWords).
		Previous(parentDictWords).
		Build()

	db.diff.NonDictionaryWords = NewDeltaBuilder().
		Current(currentNonDictWords).
		Previous(parentNonDictWords).
		Build()

	return db
}

// InformalWords adds InformalWords to the Diff schema.
func (db *DiffBuilder) InformalWords(prevInformalWords, curInformalWords []string) *DiffBuilder {
	db.diff.InformalWords = NewDeltaBuilder().
		Current(curInformalWords).
		Previous(prevInformalWords).
		Build()

	return db
}

// NonSafeWords adds NonSafeWords to the Diff schema.
func (db *DiffBuilder) NonSafeWords(prevNonSafeWords, currNonSafeWords []string) *DiffBuilder {
	db.diff.NonSafeWords = NewDeltaBuilder().
		Current(currNonSafeWords).
		Previous(prevNonSafeWords).
		Build()

	return db
}

// UpperCaseWords adds UpperCaseWords to the Diff schema.
func (db *DiffBuilder) UpperCaseWords(prevUpperCaseWords, currUpperCaseWords []string) *DiffBuilder {
	db.diff.UppercaseWords = NewDeltaBuilder().
		Current(currUpperCaseWords).
		Previous(prevUpperCaseWords).
		Build()

	return db
}

// LongestNewRepeatedCharacter determines the difference of longest repeating characters substrings
// between current and previous revisions contents.
// If the current revision content repeating substring length is greater than the previous revision string,
// add the length of the current revision. Else, add 1.
func (db *DiffBuilder) LongestNewRepeatedCharacter(currWiki, prevWiki string) *DiffBuilder {
	currLongestRepeatedChar := getLongestRepeatedCharacter(currWiki)
	prevLongestRepeatedChar := getLongestRepeatedCharacter(prevWiki)

	db.diff.LongestNewRepeatedCharacter = 1

	if currLongestRepeatedChar > prevLongestRepeatedChar {
		db.diff.LongestNewRepeatedCharacter = currLongestRepeatedChar
	}

	return db
}

// Build returns the builder's diff schema.
func (db *DiffBuilder) Build() *schema.Diff {
	return db.diff
}
