package nlp

import (
	"github.com/go-ego/gse"
	"github.com/go-ego/gse/hmm/pos"
)

var (
	seg    gse.Segmenter
	posSeg pos.Segmenter

	new, _ = gse.New("zh,testdata/test_dict3.txt", "alpha")
)

func SetCut(str string) []string {
	// 加载默认词典
	seg.LoadDict()

	hmm := new.Cut(str, true)

	return hmm
}
