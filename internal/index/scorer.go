package index

import "math"


func TF(termCount, totalTokens int) float64 {
	if totalTokens == 0 {
		return 0
	}
	return float64(termCount) / float64(totalTokens)
}

func IDF(docCount, docsWithTerm int) float64 {
	if docCount == 0 {
		return 0
	}
	return math.Log(float64(1+docCount)/float64(1+docsWithTerm)) + 1
}

func TFIDF(termCount, totalTokens, docCount, docsWithTerm int) float64 {
	return TF(termCount, totalTokens) * IDF(docCount, docsWithTerm)
}