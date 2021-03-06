package explorerlib

import (
	"github.com/coveo/go-coveo/search"
	"github.com/coveo/uabot/scenariolib"
	"math"
	"fmt"
)

type Index struct {
	Client search.Client
}

func NewIndex(endpoint string, searchToken string) (Index, error) {
	client, err := search.NewClient(search.Config{
		Endpoint:  endpoint,
		Token:     searchToken,
		UserAgent: "",
	})
	return Index{Client: client}, err
}

func (index *Index) FetchLanguages() ([]string, error) {
	languageFacetValues, err := index.Client.ListFacetValues("@syslanguage", math.MaxInt16)
	languages := []string{}
	for _, value := range languageFacetValues.Values {
		languages = append(languages, value.Value)
	}
	return languages, err
}

func (index *Index) FetchFieldValues(field string) (*search.FacetValues, error) {
	return index.Client.ListFacetValues(field, 1000)
}

func (index *Index) FindTotalCountFromQuery(query search.Query) (int, error) {
	response, status := index.Client.Query(query)
	return response.TotalCount, status
}

func (index *Index) FetchResponse(queryExpression string, numberOfResults int) (*search.Response, error) {
	return index.Client.Query(search.Query{
		AQ:              queryExpression,
		NumberOfResults: numberOfResults,
	})
}

func (index *Index) BuildGoodQueries(wordCountsByLanguage map[string]WordCounts, numberOfQueryByLanguage int, averageNumberOfWords int) (map[string][]string, error) {
	queriesInLanguage := make(map[string][]string)
	scenariolib.Info.Print("Building queries and calling the index to validate that they return results")
	for language, wordCounts := range wordCountsByLanguage {
		words := []string{}
		for i := 0; i < numberOfQueryByLanguage; {
			word := wordCounts.PickExpNWords(averageNumberOfWords)
			response, err := index.FetchResponse(word, 10)
			if err != nil {
				return nil, err
			}
			if len(response.Results) > 0 {
				words = append(words, word)
				i++
				fmt.Printf("\rBuilding and validating queries: %.0f %% completed for language %s", (float32(i)/float32(numberOfQueryByLanguage))*100, language)
			}
		}
		fmt.Printf("\n")
		scenariolib.Info.Println("Language ", language," : Total number of good queries : ", len(words))
		queriesInLanguage[language] = words

	}
	fmt.Printf("\n")
	return queriesInLanguage, nil
}
