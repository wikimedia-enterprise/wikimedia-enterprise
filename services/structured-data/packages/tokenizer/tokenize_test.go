package tokenizer_test

import (
	"testing"
	"wikimedia-enterprise/services/structured-data/packages/tokenizer"

	"github.com/stretchr/testify/suite"
)

type tokenizeTestSuite struct {
	suite.Suite
	val    string
	result []string
}

// TestTokenize method tests Tokenize correct behaviour.
func (s *tokenizeTestSuite) TestTokenize() {
	s.Assert().Equal(s.result, tokenizer.Tokenize(s.val))
}

// TestTokenize test runner.
func TestTokenize(t *testing.T) {
	for _, testCase := range []*tokenizeTestSuite{
		{
			val: `That's that :: :) > n95 Latin. follows 3rd: 2nd m80. 
				we'll follows: https://en.wikipedia.org/wiki/English_Wikipedia total- \"Oh\" 43 ँ هذا هو :: :)>
				ありがとうございます この項目では、をご覧ください。`,
			result: []string{
				"That's", "that", "n95", "Latin", "follows", "3rd", "2nd", "m80", "we'll",
				"follows", "total", "Oh", "ँ", "هذا", "هو", "ありがとうございます", "この項目では", "をご覧ください",
			},
		},
		{
			val: `{{\nshort description|'Form' of communication for marketing, that's typically paid for}}
				<ref name=\"Stanton\">William J. Stanton. ''Fundamentals of Marketing''. McGraw-Hill (1984).</ref>
				[File:Cocacola-5cents-1900 edit1.jpg|thumb|A [[Coca-Cola]] advertisement from the 1890s]]\n`,
			result: []string{
				"short", "description", "Form", "of", "communication", "for", "marketing",
				"that's", "typically", "paid", "for", "William", "J", "Stanton", "Fundamentals",
				"of", "Marketing", "McGraw", "Hill", "File", "Cocacola", "5cents",
				"edit1", "jpg", "thumb", "A", "Coca", "Cola", "advertisement", "from", "the", "1890s",
			},
		},
		{
			val: `{Dablink|ありがとうございます この項目では、<u>ユーラシア大陸の東部地域を指す、
				「中国」という用語の意味と呼称の変遷について説明</u>しています。\n* 現代的な定義の
				[[国家]]および政権については「[[中華人民共和国]]」をご覧ください。\n}}
				gwo<sup>2</sup>\n\n[[ベトナム]]`,
			result: []string{
				"Dablink", "ありがとうございます", "この項目では", "ユーラシア大陸の東部地域を指す", "中国",
				"という用語の意味と呼称の変遷について説明", "しています", "現代的な定義の", "国家", "および政権については",
				"中華人民共和国", "をご覧ください", "gwo", "ベトナム",
			},
		},
		{
			val: `			{{تسويق}}\n[[ملف:Cocacola-5cents-1900 edit1.jpg|تصغير|يسار|200px|اعلان
				[[كوكاكولا]] من عام 1890]]\n[[ملف:Chipsy Announcement.jpg|تصغير|إعلان لشركة
				[[شيبسي]] المصرية على جانب [[طريق سريع]].]]\n`,
			result: []string{
				"تسويق", "ملف", "Cocacola", "5cents", "edit1", "jpg", "تصغير", "يسار", "200px", "اعلان", "كوكاكولا", "من",
				"عام", "ملف", "Chipsy", "Announcement", "jpg", "تصغير", "إعلان", "لشركة", "شيبسي", "المصرية", "على", "جانب", "طريق", "سريع",
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
