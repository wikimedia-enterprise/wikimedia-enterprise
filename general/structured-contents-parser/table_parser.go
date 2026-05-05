// Package parser implements utility routines for traversing and retrieving tables from a page html.
// The parser packages adds a HTML table parsing utility to the goquery package.

package parser

import (
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"math"
	"slices"
	"sort"
	"wikimedia-enterprise/general/parser/utils"

	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/araddon/dateparse"
)

var (
	unwantedClasses = []string{
		"sidebar", "infobox", "navbox", "vcard", "navigation", "hatnote",
	}
	unwantedRole = "navigation"
)

// ToDo Remove in future, this was for Election 2024 POC
// FindHeadingAndReturnTables finds a heading (h1-h6) that contains a keyword string, then moves to following siblings to find <table> tags.
func (p *Parser) GetTablesFromHeading(sel *goquery.Selection, headingTag string, headingText string, mode ValMode) []TableWithHeaders {
	tbs := []TableWithHeaders{}

	sel.Find(headingTag). // Get the first header (h1-h6)
				FilterFunction(filterByKeyword(headingText)).    // Filter the header by the keyword
				First().                                         // Get the first header that contains the keyword
				NextFilteredUntil("table", "h1,h2,h3,h4,h5,h6"). // Get all following siblings, until the next header h1-h6
				Each(func(j int, tbl *goquery.Selection) {
			table, _ := p.GetTableRows(tbl, mode)
			if tableHasRows(table) {
				tbs = append(tbs, table)
			}
		})

	return tbs
}

// ToDo Remove in future, this was for Election 2024 POC
// filterByKeyword returns a function that filters a selection by a keyword.
func filterByKeyword(keyword string) func(int, *goquery.Selection) bool {
	return func(i int, s *goquery.Selection) bool {
		return strings.Contains(s.Text(), keyword)
	}
}

// hasUnwantedAttributes checks if a selection or any of its ancestors has unwanted classes or roles
func hasUnwantedAttributes(s *goquery.Selection) bool {
	// Check current element
	if containsUnwanted(s) {
		return true
	}

	// Check ancestors using Closest
	if s.Closest(fmt.Sprintf("%s, *[role='%s']", unwantedClasses, unwantedRole)).Length() > 0 {
		return true
	}

	return s == nil
}

// containsUnwanted checks a single selection for unwanted classes or roles
func containsUnwanted(s *goquery.Selection) bool {
	classes := strings.Fields(s.AttrOr("class", ""))
	for _, class := range classes {
		for _, pattern := range unwantedClasses {
			if strings.Contains(class, pattern) {
				return true
			}
		}
	}
	role := s.AttrOr("role", "")
	return strings.Contains(role, "navigation")
}

// GetTables finds a section that contains a keyword string, then moves to following siblings to find <table> tags.
func (p *Parser) GetTables(sel *goquery.Selection, mode ValMode) []TableWithHeaders {
	tables := []TableWithHeaders{}
	sel.FindMatcher(utils.TopOnly("table")).Each(func(i int, s *goquery.Selection) {
		// Skip table if it or any of its ancestors have unwanted attributes
		if hasUnwantedAttributes(s) {
			return
		}

		table, _ := p.GetTableRows(s, mode)
		if tableHasRows(table) {
			tables = append(tables, table)
		}
	})

	return tables
}

// GetTableRows takes an HTML table and returns a DataFrame with basic type conversions.
// Note that the provided selection must contain a <table> element at the top level.
func (p *Parser) GetTableRows(sel *goquery.Selection, mode ValMode) (TableWithHeaders, error) {
	var headers, records [][]Cell
	var cells []Cell
	var errs []error

	// Skip unwanted tables by class or role
	if hasUnwantedAttributes(sel) {
		return TableWithHeaders{}, nil // not an error, just skip
	}

	// We use state and current to keep track of the rowspan states for each column
	state := make(map[int]SpanCell)
	cur := make(map[int]SpanCell)

	tableHTML, _ := sel.Html()

	// Prepare the table by removing sup and tfoot elements, we want to ignore these HTML elements
	slc := p.prepareTable(sel)

	// Get the column count and check if the table is balanced
	confidenceScore := p.getTableBalanceConfidence(slc)

	// Iterate over each row in the table,
	slc.FindMatcher(utils.FindIgnoring("tr", "table")).Each(func(_ int, tr *goquery.Selection) {
		var err error
		cells, cur, err = p.parseDataRow(tr, state, mode, tableHTML)
		if err != nil {
			errs = append(errs, err)
		}
		isEmptyRow := true
		for _, cell := range cells {
			if strings.TrimSpace(cell.Value) != "" || len(cell.Inner) > 0 {
				isEmptyRow = false
				break
			}
		}
		if isEmptyRow {
			return // skip to next row
		}
		// If the row is a mix of header and data cells, with mostly header cells then we parse it as a header row
		if tr.FindMatcher(utils.FindIgnoring("th", "table")).Length() > tr.FindMatcher(utils.FindIgnoring("td", "table")).Length() && len(records) == 0 {
			if len(cells) > 0 {
				headers = append(headers, cells)
			}
		} else {
			if len(cells) > 0 {
				records = append(records, cells)
			}
		}

		state = p.mergeState(state, cur)
	})

	// Pad the records and headers to the same length
	padImbalancedRows(records, headers)

	// Compose error
	var err error
	if len(errs) == 1 {
		err = errs[0]
	} else if len(errs) > 1 {
		// Go 1.20+: errors.Join(errs...) is best, otherwise do this:
		err = fmt.Errorf("multiple errors in table row parsing: %v", errs)
	}

	return TableWithHeaders{
		Headers:         headers,
		Rows:            records,
		ConfidenceScore: confidenceScore,
	}, err
}

// prepareTable removes sup, tfoot, figure, and gallery elements from the table
func (p *Parser) prepareTable(sel *goquery.Selection) *goquery.Selection {
	slc := sel.Clone()

	// Remove sup, tfoot, figure, gallery and other elements
	slc.FindMatcher(utils.TopOnly("sup, tfoot, style, link, figure, .gallery, .gallerybox, .mw-gallery-packed")).Remove()

	// Replace <br> with space
	slc.Find("br").Each(func(index int, item *goquery.Selection) {
		item.ReplaceWithHtml(" ")
	})

	// Replace a HTML list with a newline after the simplified plain text list
	slc.Find("ul, ol, dl").Each(func(_ int, list *goquery.Selection) {
		list.Find("li, dt, dd").Each(func(_ int, item *goquery.Selection) {
			// Replace list items with newline and the text in the list item
			item.ReplaceWithHtml(fmt.Sprintf("\n%s", item.Text()))
		})

		list.ReplaceWithHtml(fmt.Sprintf("%s\n", list.Text()))
	})

	return slc
}

// parseDataRow returns the data of a table row and a map of the next row's rowspan states
func (p *Parser) parseDataRow(
	row *goquery.Selection,
	state map[int]SpanCell,
	mode ValMode,
	tableHTML string,
) ([]Cell, map[int]SpanCell, error) {
	var rec []Cell                 // Slice to store the current row's data
	var val string                 // The value of the placeholder for rowspan and colspan cells
	next := make(map[int]SpanCell) // Map to track the next row's rowspan states
	i := 0                         // The `cell index` in the current row
	var errs []error

	row.FindMatcher(utils.FindIgnoring("th, td", "table")).Each(func(_ int, cell *goquery.Selection) {
		var txt string // The text of the cell
		var inner []TableWithHeaders

		// Get the colspan and rowspan attributes of the current cell
		colSpan, rowSpan := getSpans(cell)

		// add cells carried over by rowspan from `state` for the current row and cell index `i`
		var err error
		rec, err = p.addCellsforState(state, rec, i, tableHTML)
		if err != nil {
			errs = append(errs, err)
		}

		// If the cell does not contain a table, we parse the cell data
		if cell.FindMatcher(utils.TopOnly("table")).Length() == 0 {
			// Convert the cell data to a basic type and remove extra spaces
			txt = p.convertDataType(spaces.ReplaceAllString(cell.Text(), " "))
		} else {
			// If the cell has a inner table, we parse the table and add it to the current row
			// Recursively parse the inner table and add them to a `Inner` field in the Cell struct
			inner = p.GetTables(cell, mode)
			// The cell could have text as well as the inner Table. Clone the cell to remove any nested tables
			cln := cell.Clone()
			cln.FindMatcher(utils.TopOnly("table")).Remove()
			txt = p.convertDataType(spaces.ReplaceAllString(cln.Text(), " ")) //ToDo, should we output plain text if it has an inner table?
		}

		// Keep track of next colspans, store both rowspan and colspan for multi cell spans (greedy cells with 2x2, 2x3, 1x4, etc.)
		if rowSpan > 1 {
			switch mode { //ToDo: PMs and client feedback to decide what way we output extra spanned cells
			case Duplicate:
				val = txt
			case Empty:
				val = ""
			case Placeholder:
				val = "____"
			case Zero:
				val = "0"
			}

			pos := len(rec) // The position of the current cell in the row, after the previous colspan added
			// Add the cell to the nextRowSpan map, with the rowspan decremented by 1
			next[pos] = SpanCell{RowSpanRemaining: rowSpan - 1, Value: val, ColSpanRemaining: colSpan} //Ignore inner Tables, do not duplicate if inner table has rowspan
		}

		cellData := Cell{Value: txt, Inner: inner}
		// Add cells to row record, if cell has colspan create many cells
		rec = p.addCellsForColSpan(colSpan, cellData, rec)

		// Move to the next cell index
		i = len(rec)
	})

	// add cells carried over by rowspan from `state`, we add the cells after the current cell index
	var err error
	rec, err = p.addCellsforState(state, rec, i, tableHTML)
	if err != nil {
		errs = append(errs, err)
	}

	// Combine all errors, if any
	var allErr error
	if len(errs) == 1 {
		allErr = errs[0]
	} else if len(errs) > 1 {
		// Use errors.Join if Go 1.20+, otherwise make your own
		allErr = fmt.Errorf("multiple errors in parseDataRow: %v", errs)
	}

	return rec, next, allErr
}

// addCellsForColSpan adds cells to the record based on the colspan attribute
func (p *Parser) addCellsForColSpan(csp int, data Cell, rec []Cell) []Cell {
	for j := 0; j < csp; j++ {
		if j == 0 {
			rec = append(rec, data)
		} else {
			rec = append(rec, Cell{Value: data.Value}) // Do not duplicate inner tables
		}
	}

	return rec
}

// addCellsforState adds cells to the record based on the rowspan and colspan
func (p *Parser) addCellsforState(
	state map[int]SpanCell,
	rec []Cell,
	i int,
	tableHTML string,
) ([]Cell, error) {
	const maxIterations = 1000
	const warnThreshold = 950
	count := 0
	var errs []error

	for {
		cur, ok := state[i]
		if !ok {
			errs = append(errs, fmt.Errorf("addCellsforState: no cell at state[%d] (normal break)", i))
			break
		}
		if cur.RowSpanRemaining <= 0 {
			errs = append(errs, fmt.Errorf("addCellsforState: RowSpanRemaining exhausted at state[%d]", i))
			break
		}
		if count >= maxIterations {
			errs = append(errs, fmt.Errorf(
				"addCellsforState: hit maxIterations (%d) for table: '%s'... State: %+v",
				maxIterations, truncateString(tableHTML, 2000), state,
			))
			break
		}

		rec = p.addCellsForColSpan(cur.ColSpanRemaining, Cell{Value: cur.Value}, rec)
		i += cur.ColSpanRemaining
		count++
		if count == warnThreshold {
			errs = append(errs, fmt.Errorf(
				"addCellsforState: approaching maxIterations (%d) for table: '%s'... State: %+v",
				maxIterations, truncateString(tableHTML, 2000), state,
			))
		}
	}

	if len(errs) == 0 {
		return rec, nil
	}
	if len(errs) == 1 {
		return rec, errs[0]
	}
	return rec, fmt.Errorf("multiple errors: %v", errs)
}

// Helper to avoid flooding logs with huge HTML
func truncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen] + "...(truncated)"
}

// getSpans returns the colspan and rowspan attributes of a cell
func getSpans(c *goquery.Selection) (int, int) {
	csp, _ := strconv.Atoi(c.AttrOr("colspan", "1"))
	rsp, _ := strconv.Atoi(c.AttrOr("rowspan", "1"))
	// Default to 1 if invalid
	if csp < 1 {
		csp = 1 // safeguard
	}
	if rsp < 1 {
		rsp = 1
	}
	// https://html.spec.whatwg.org/#the-td-element
	// Max colspan is 1000
	// Max rowspan is 65534, but there's no point supporting such numbers
	if csp > 1000 {
		log.Printf("excessive colspan detected: %d - using 1000\n", csp)
		csp = 1000
	}
	if rsp > 1000 {
		log.Printf("excessive rowspan detected: %d - using 1000\n", rsp)
		rsp = 1000
	}

	return csp, rsp
}

// mergeState combines the current and next rowspan states for the next row.
func (p *Parser) mergeState(state, next map[int]SpanCell) map[int]SpanCell {
	// Decrease the row count in the `state`` and remove them if their span is completed, ie. Rows == 1
	for idx, span := range state {
		if span.RowSpanRemaining > 1 {
			state[idx] = SpanCell{RowSpanRemaining: span.RowSpanRemaining - 1, Value: span.Value, ColSpanRemaining: span.ColSpanRemaining}
		} else {
			delete(state, idx)
		}
	}

	// Add or update spans from nextRowSpan into currentRowSpan
	for idx, span := range next {
		// If there is already a span in place, we only update it if the new span is longer, but it's a safeguard for invalid HTML table tags by editors
		if existing, exists := state[idx]; exists {
			if span.RowSpanRemaining > existing.RowSpanRemaining {
				state[idx] = span
			}
		} else {
			state[idx] = span
		}
	}

	return state
}

// convertDataType detects and converts numeric values, percentages, dates, booleans, etc.
func (p *Parser) convertDataType(txt string) string {
	var val string
	txt = strings.TrimSpace(txt)

	if _, isNumeric := parseNumericValue(txt); isNumeric {
		return txt // <-- return the original text
	} else if dte, isDate := parseDateValue(txt); isDate {
		if dte.Year() == 0 {
			val = txt
		} else {
			val = fmt.Sprintf("%d-%02d-%02d", dte.Year(), dte.Month(), dte.Day())
		}
	} else {
		val = txt
	}

	return val
}

// padImbalancedRows pads the records and headers to the same length
func padImbalancedRows(records [][]Cell, headers [][]Cell) {
	maxCols := 0
	for _, record := range records {
		if len(record) > maxCols {
			maxCols = len(record)
		}
	}

	for _, header := range headers {
		if len(header) > maxCols {
			maxCols = len(header)
		}
	}

	for i, record := range records {
		if len(record) < maxCols {
			records[i] = padRecordToLength(record, maxCols)
		}
	}

	for i, header := range headers {
		if len(header) < maxCols {
			headers[i] = padRecordToLength(header, maxCols)
		}
	}
}

// padRecordToLength is a helper function to pad a record or header to a specific length
func padRecordToLength(record []Cell, length int) []Cell {
	for len(record) < length {
		record = append(record, Cell{Value: ""})
	}

	return record
}

func (p *Parser) getTableBalanceConfidence(sel *goquery.Selection) float64 {
	var (
		colCounts           []int
		rowspans            = map[int]int{}
		seenRows            = map[uint64]bool{}
		headerColCount      int
		headerDetected      bool
		deviationPenalty    = 1.0
		compoundHeaderCount int
		nonEmptyCells       int
		totalCells          int
	)

	// Normalize all th/td cell text to remove unicode quirks
	normalizeUnicode(sel)

	rows := sel.FindMatcher(utils.TopOnly("tr"))

	// Special case: single celled table.
	if rows.Length() == 1 && rows.FindMatcher(utils.TopOnly("td,th")).Length() == 1 {
		return 1.0
	}

	rows.Each(func(rowIndex int, rowSel *goquery.Selection) {
		text := strings.TrimSpace(rowSel.Text())
		if text == "" {
			return // skip visually empty row
		}

		// Skip duplicate rows
		rowHash := hashRow(rowSel)
		if seenRows[rowHash] {
			return
		}
		seenRows[rowHash] = true

		// Skip long single-cell metadata rows (e.g., summary/notes)
		if isSummaryRow(rowSel) {
			return
		}

		curColCount := 0

		// Add in any colspan from active rowspans
		if spanCount, ok := rowspans[rowIndex]; ok {
			curColCount += spanCount
			delete(rowspans, rowIndex)
		}

		rowSel.FindMatcher(utils.TopOnly("th, td")).Each(func(_ int, cellSel *goquery.Selection) {
			text := strings.TrimSpace(normalizeCellContent(cellSel))
			if text != "" && text != "–" && text != "—" && text != "-" {
				nonEmptyCells++
			}
			totalCells++

			colspan, rowspan := getSpans(cellSel)
			if colspan == 0 {
				colspan = 1
			}
			if rowspan == 0 {
				rowspan = 1
			}
			curColCount += colspan

			// Track rowspan contributions
			for i := 1; i < rowspan; i++ {
				rowspans[rowIndex+i] += colspan
			}
		})

		if curColCount == 0 {
			return // ignore fully empty row
		}

		colCounts = append(colCounts, curColCount)

		// Detect header structure for comparison
		if !headerDetected && rowSel.FindMatcher(utils.TopOnly("th")).Length() > 0 {
			headerColCount = curColCount
			headerDetected = true
		}

		// Detect repeated headers (compound table detection)
		if headerDetected && rowSel.FindMatcher(utils.TopOnly("th")).Length() > 0 && curColCount == headerColCount {
			compoundHeaderCount++
		}
	})

	if len(colCounts) == 0 {
		return 0.0
	}
	if len(colCounts) == 1 {
		return 1.0
	}

	// Special case: all rows are clean 2-column rows with 1 <th> and 1 <td>
	if headerColCount == 2 && allRowsHaveExactlyOneThAndOneTd(sel) {
		return 1.0
	}

	// Compute median to stabilize std deviation
	median := computeMedian(colCounts)

	// Compute standard deviation (Bessel's correction)
	var sumSq float64
	for _, count := range colCounts {
		diff := float64(count) - float64(median)
		sumSq += diff * diff
	}
	stdDev := math.Sqrt(sumSq / float64(len(colCounts)-1))

	// Apply soft penalty if most rows diverge from *median* (not header)
	mismatchCount := 0
	for _, count := range colCounts {
		if count != median {
			mismatchCount++
		}
	}

	// Handle trailing empty columns (likely structural, not semantic)
	if headerDetected && headerColCount > 0 {
		consistentlyEmptyTrailing := 0
		for i := headerColCount - 1; i >= 0; i-- {
			allEmpty := true
			sel.FindMatcher(utils.TopOnly("tr")).Each(func(_ int, tr *goquery.Selection) {
				cells := tr.FindMatcher(utils.TopOnly("td, th"))
				if cells.Length() <= i {
					return
				}
				text := strings.TrimSpace(cells.Eq(i).Text())
				if text != "" && text != "–" && text != "—" && text != "-" {
					allEmpty = false
				}
			})
			if allEmpty {
				consistentlyEmptyTrailing++
			} else {
				break
			}
		}
		if consistentlyEmptyTrailing > 0 {
			deviationPenalty = 1.0
		}
	}

	if float64(mismatchCount)/float64(len(colCounts)) > 0.2 {
		deviationPenalty *= 0.8 // mild penalty
	}

	// Penalize if multiple compound headers detected
	if compoundHeaderCount > 1 {
		deviationPenalty *= 0.8
	}

	// Special override: all rows are consistently underfilled but balanced
	if stdDev < 0.01 && headerDetected && headerColCount > 0 {
		mostCommon := mode(colCounts)
		if mostCommon < headerColCount {
			deviationPenalty = 1.0 // neutralize mismatch penalty
		}
	}

	// Bonus override for sparse but visually balanced tables
	emptyFraction := 1.0
	if totalCells > 0 {
		emptyFraction = float64(nonEmptyCells) / float64(totalCells)
	}
	if emptyFraction < 0.5 && stdDev < 0.05 {
		deviationPenalty = 1.0
	}

	// Compute grid completeness bonus
	gridCompleteness := computeGridCompleteness(sel)

	// Final confidence score
	confidence := (1.0 / (1.0 + stdDev)) * deviationPenalty

	// Apply grid bonus if table is complete and not penalized
	if stdDev < 0.02 && gridCompleteness > 0.95 {
		confidence += 0.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}
	return confidence
}

func mode(counts []int) int {
	freq := map[int]int{}
	maxCount := 0
	modeVal := counts[0]
	for _, v := range counts {
		freq[v]++
		if freq[v] > maxCount {
			maxCount = freq[v]
			modeVal = v
		}
	}
	return modeVal
}

func computeGridCompleteness(sel *goquery.Selection) float64 {
	var (
		totalCells    int
		nonEmptyCells int
	)

	sel.FindMatcher(utils.TopOnly("tr")).Each(func(_ int, rowSel *goquery.Selection) {
		rowSel.FindMatcher(utils.TopOnly("th, td")).Each(func(_ int, cellSel *goquery.Selection) {
			colspan, _ := getSpans(cellSel)
			if colspan == 0 {
				colspan = 1
			}
			cellText := strings.TrimSpace(cellSel.Text())
			totalCells += colspan
			if cellText != "" {
				nonEmptyCells += colspan
			}
		})
	})

	if totalCells == 0 {
		return 0.0
	}
	return float64(nonEmptyCells) / float64(totalCells)
}

func anyRowNonEmpty(row []Cell) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell.Value) != "" {
			return true
		}
	}
	return false
}

func normalizeCellContent(cellSel *goquery.Selection) string {
	text := strings.TrimSpace(cellSel.Text())
	text = strings.ReplaceAll(text, "\u200D", "") // remove ZWJ
	text = strings.ReplaceAll(text, "\u200C", "") // remove ZWNJ
	text = strings.ReplaceAll(text, "–", "-")     // normalize en dash
	text = strings.ReplaceAll(text, "—", "-")     // normalize em dash
	return text
}

func allRowsHaveExactlyOneThAndOneTd(sel *goquery.Selection) bool {
	valid := true
	sel.FindMatcher(utils.TopOnly("tr")).EachWithBreak(func(_ int, rowSel *goquery.Selection) bool {
		// skip empty or summary rows
		text := strings.TrimSpace(rowSel.Text())
		if text == "" || isSummaryRow(rowSel) {
			return true
		}

		if rowSel.FindMatcher(utils.TopOnly("th")).Length() != 1 || rowSel.FindMatcher(utils.TopOnly("td")).Length() != 1 {
			valid = false
			return false // break
		}
		return true
	})
	return valid
}

func normalizeUnicode(sel *goquery.Selection) {
	sel.FindMatcher(utils.TopOnly("th, td")).Each(func(_ int, cellSel *goquery.Selection) {
		node := cellSel.Get(0)
		if node == nil || node.FirstChild == nil {
			return
		}
		text := node.FirstChild.Data
		if text == "" {
			return
		}
		normalized := normalizeText(text)
		node.FirstChild.Data = normalized
	})
}

func normalizeText(s string) string {
	// Replace common problematic characters with normalized equivalents
	replacements := map[rune]string{
		'\u2013': "-", // en dash
		'\u2014': "-", // em dash
		'\u200D': "",  // zero-width joiner
		'\u200C': "",  // zero-width non-joiner
		'\u00A0': " ", // non-breaking space
		'\u202F': " ", // narrow non-breaking space
		'\uFEFF': "",  // BOM / zero-width space
	}

	var b strings.Builder
	for _, r := range s {
		if repl, ok := replacements[r]; ok {
			b.WriteString(repl)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// parseNumericValue tries to parse a string into a float64 or int, handling percentages.
// It returns an interface{} that must be type asserted by the caller to either float64 or int.
func parseNumericValue(value string) (interface{}, bool) {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\u00A0", "") // remove non-breaking space
	value = strings.ReplaceAll(value, " ", "")      // remove normal spaces

	// Handle percentages directly
	percent := percentPattern.FindStringSubmatch(value)
	if len(percent) > 0 {
		val := percent[1]

		if num, err := strconv.ParseFloat(val, 64); err == nil {
			return num / 100, true
		}
	}

	// Now parse the value as-is, no replacing anything
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		if float64(int(num)) == num {
			return int(num), true
		}
		return num, true
	}

	return nil, false

}

// parseDateValue tries to parse any given date string into a Go time.Time object.
func parseDateValue(dateString string) (time.Time, bool) {
	t, err := dateparse.ParseAny(dateString)
	if err != nil {
		return time.Time{}, false // Return the zero value of time.Time on error
	}

	return t, true
}

func tableHasRows(tbl TableWithHeaders) bool {
	return slices.ContainsFunc(tbl.Rows, anyRowNonEmpty)
}

func computeMedian(nums []int) int {
	sort.Ints(nums)
	n := len(nums)
	if n%2 == 1 {
		return nums[n/2]
	}
	return (nums[n/2-1] + nums[n/2]) / 2
}

func isSummaryRow(sel *goquery.Selection) bool {
	cells := sel.FindMatcher(utils.TopOnly("td, th"))
	if cells.Length() != 1 {
		return false
	}

	colspan, _ := strconv.Atoi(cells.First().AttrOr("colspan", "1"))
	content := strings.TrimSpace(cells.First().Text())
	if colspan >= 5 || len(content) > 100 {
		return true
	}
	return false
}

func hashRow(sel *goquery.Selection) uint64 {
	h := fnv.New64a()
	sel.FindMatcher(utils.TopOnly("td, th")).Each(func(i int, cell *goquery.Selection) {
		text := normalize(cell.Text())
		_, _ = io.WriteString(h, text+"|")
	})
	return h.Sum64()
}

func normalize(text string) string {
	text = strings.ReplaceAll(text, "–", "-")     // en dash
	text = strings.ReplaceAll(text, "\u200D", "") // ZWJ
	text = strings.ReplaceAll(text, "\u200C", "") // ZWNJ
	text = strings.TrimSpace(text)
	return text
}
