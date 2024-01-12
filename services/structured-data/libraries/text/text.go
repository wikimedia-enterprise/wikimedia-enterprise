package text

import (
	"context"
	"errors"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/packages/textprocessor"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// New creates new instance of WordsPairGetter for dependency injection.
func New(env *env.Environment) (WordsPairGetter, error) {
	cnn, err := grpc.Dial(env.TextProcessorURL, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	txt := &Text{
		Processor: textprocessor.NewTextProcessorClient(cnn),
	}

	return txt, nil
}

// Words represents aggregated text processor API response.
type Words struct {
	Dictionary    []string
	NonDictionary []string
	Informal      []string
	NonSafe       []string
	UpperCase     []string
}

// WordsGetter interface representation for GetWords method for dependency injection.
type WordsGetter interface {
	GetWords(ctx context.Context, tks []string) (*Words, error)
}

// WordsPairGetter interface representation for WordsPairGetter method for dependency injection.
type WordsPairGetter interface {
	GetWordsPair(ctx context.Context, cts, prt []string) (*Words, *Words, error)
}

// Text wrapper for text processor service.
type Text struct {
	Processor textprocessor.TextProcessorClient
}

// GetWords runs set of tokens against text processor service.
func (t *Text) GetWords(ctx context.Context, tks []string) (*Words, error) {
	nrs := 4
	ers := make(chan error, nrs)
	dws := make(chan *textprocessor.DictionaryResponse, 1)
	iws := make(chan *textprocessor.InformalResponse, 1)
	nws := make(chan *textprocessor.NonSafeResponse, 1)
	uws := make(chan *textprocessor.UppercaseResponse, 1)

	go func() {
		res, err := t.Processor.GetDictionaryWords(ctx, &textprocessor.DictionaryRequest{
			Tokens: tks,
		})
		dws <- res
		ers <- err
	}()

	go func() {
		res, err := t.Processor.GetInformalWords(ctx, &textprocessor.InformalRequest{
			Tokens: tks,
		})
		iws <- res
		ers <- err
	}()

	go func() {
		res, err := t.Processor.GetNonSafeWords(ctx, &textprocessor.NonSafeRequest{
			Tokens: tks,
		})
		nws <- res
		ers <- err
	}()

	go func() {
		res, err := t.Processor.GetUpperCaseWords(ctx, &textprocessor.UppercaseRequest{
			Tokens: tks,
		})
		uws <- res
		ers <- err
	}()

	ter := ""

	for i := 0; i < nrs; i++ {
		if err := <-ers; err != nil {
			ter += err.Error()
		}
	}

	if len(ter) != 0 {
		return nil, errors.New(ter)
	}

	dwp := (<-dws)
	wds := &Words{
		Dictionary:    dwp.GetDictWords(),
		NonDictionary: dwp.GetNonDictWords(),
		Informal:      (<-iws).GetInformalWords(),
		NonSafe:       (<-nws).GetNonSafeWords(),
		UpperCase:     (<-uws).GetUppercaseWords(),
	}

	return wds, nil
}

// GetWordsPair runs two sets of tokens against text processor service.
func (t *Text) GetWordsPair(ctx context.Context, cts, pts []string) (*Words, *Words, error) {
	nrs := 2
	ers := make(chan error, nrs)
	cws := make(chan *Words, 1)
	pws := make(chan *Words, 1)

	go func() {
		res, err := t.GetWords(ctx, cts)
		cws <- res
		ers <- err
	}()

	go func() {
		res, err := t.GetWords(ctx, pts)
		pws <- res
		ers <- err
	}()

	ter := ""

	for i := 0; i < nrs; i++ {
		if err := <-ers; err != nil {
			ter += err.Error()
		}
	}

	if len(ter) != 0 {
		return nil, nil, errors.New(ter)
	}

	return <-cws, <-pws, nil
}
