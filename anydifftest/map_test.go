package anydifftest

import (
	"testing"

	"github.com/unixpickle/anydiff"
	"github.com/unixpickle/anydiff/anyseq"
	"github.com/unixpickle/anyvec"
)

func TestMap(t *testing.T) {
	runWithCreators(t, func(t *testing.T, c anyvec.Creator, prec float64) {
		inSeq, varList := makeBasicTestSeqs(c)
		t.Run("SameSize", func(t *testing.T) {
			ch := &SeqChecker{
				F: func() anyseq.Seq {
					return anyseq.Map(inSeq, func(v anydiff.Res, n int) anydiff.Res {
						return anydiff.Tanh(v)
					})
				},
				V: varList,
			}
			ch.FullCheck(t)
		})
		t.Run("DiffSize", func(t *testing.T) {
			combVar := makeRandomVec(c, 6)
			combiner := &anydiff.Matrix{
				Data: combVar,
				Rows: 6,
				Cols: 1,
			}
			ch := &SeqChecker{
				F: func() anyseq.Seq {
					return anyseq.Map(inSeq, func(v anydiff.Res, n int) anydiff.Res {
						v = anydiff.Tanh(v)
						mat1 := &anydiff.Matrix{
							Data: v,
							Rows: n,
							Cols: 6,
						}
						return anydiff.MatMul(false, false, mat1, combiner).Data
					})
				},
				V: append([]*anydiff.Var{combVar}, varList...),
			}
			ch.FullCheck(t)
		})
	})
}

func TestMapN(t *testing.T) {
	runWithCreators(t, func(t *testing.T, c anyvec.Creator, prec float64) {
		inSeq, varList := makeBasicTestSeqs(c)

		combVar := makeRandomVec(c, 12)
		combiner := &anydiff.Matrix{
			Data: combVar,
			Rows: 2,
			Cols: 6,
		}
		varList = append(varList, combVar)

		ch := &SeqChecker{
			F: func() anyseq.Seq {
				reducedSeq := anyseq.Map(inSeq, func(v anydiff.Res, n int) anydiff.Res {
					v = anydiff.Tanh(v)
					mat1 := &anydiff.Matrix{
						Data: v,
						Rows: n,
						Cols: 6,
					}
					return anydiff.MatMul(false, true, combiner, mat1).Data
				})
				return anyseq.MapN(func(n int, v ...anydiff.Res) anydiff.Res {
					return anydiff.ScaleRepeated(v[1], v[0])
				}, reducedSeq, inSeq)
			},
			V:     varList,
			Prec:  prec * 2,
			Delta: prec,
		}
		ch.FullCheck(t)
	})
}
