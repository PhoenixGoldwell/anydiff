package anydifftest

import (
	"fmt"
	"testing"

	"github.com/unixpickle/anydiff"
	"github.com/unixpickle/anyvec"
)

var (
	testMat2x3 = []float64{
		0.437363817383076, 0.403871997911220, -1.049530750558118,
		-0.331470568323395, 0.116345399360686, 0.511644461073244,
	}
	testMat3x4 = []float64{
		0.3562578410664004, 0.7578617982451722, -0.1913118498264184, 1.8452133159741528,
		-0.8814597466441325, 0.9602627040929411, -0.0130910013786673, 0.0322210430457061,
		0.3242423917388176, -1.2307395594558537, -0.1558666924323481, 0.4826032047570410,
	}
	testMat2x4 = []float64{
		1.852319956790981, 1.385962232018235, -0.849984774797785, -0.583428448655087,
		-1.646747420778373, -1.196860898000198, 0.787446463442505, -1.774956170059281,
	}
)

type matMulExpected struct {
	Rows int
	Cols int
	Data []float64
}

func TestMatMulOut(t *testing.T) {
	runWithCreators(t, func(t *testing.T, c anyvec.Creator, prec float64) {
		m2x3 := makeMatrix(c, testMat2x3, 2, 3)
		m3x4 := makeMatrix(c, testMat3x4, 3, 4)
		m2x4 := makeMatrix(c, testMat2x4, 2, 4)

		expected := []matMulExpected{
			{
				Rows: 2,
				Cols: 4,
				Data: []float64{
					-0.5404849803784302, 2.0109835595282268, 0.0746269168633727, 0.3135358130797813,
					-0.0547459515133194, -0.7691878117953630, -0.0178571600432873, -0.3609638797927409,
				},
			},
			{
				Rows: 3,
				Cols: 2,
				Data: []float64{
					0.7963335554421418, -4.9195424331402533,
					-0.3095291604278108, 0.2347412800711330,
					-1.2542380174029129, -0.0402574796664327,
				},
			},
			{
				Rows: 3,
				Cols: 4,
				Data: []float64{
					1.355986030767449, 1.002893894608438, -0.632767912584587, 0.333175236964991,
					0.556508675403358, 0.420502076517676, -0.251669275927321, -0.442138397649835,
					-2.786615951152524, -2.066977230752253, 1.294977780068564, -0.295820395444677,
				},
			},
			{
				Rows: 4,
				Cols: 2,
				Data: []float64{
					-0.5404849803784301, -0.0547459515133194,
					2.0109835595282268, -0.7691878117953630,
					0.0746269168633727, -0.0178571600432873,
					0.3135358130797812, -0.3609638797927409,
				},
			},
		}
		actuals := []*anydiff.Matrix{
			anydiff.MatMul(false, false, m2x3, m3x4),
			anydiff.MatMul(false, true, m3x4, m2x4),
			anydiff.MatMul(true, false, m2x3, m2x4),
			anydiff.MatMul(true, true, m3x4, m2x3),
		}

		for i, e := range expected {
			a := actuals[i]
			if a.Rows != e.Rows || a.Cols != e.Cols {
				t.Errorf("mul %d: expected %dx%d but got %dx%d", i, e.Rows, e.Cols,
					a.Rows, a.Cols)
			}
			aData := getComponents(a.Data.Output())
			eData := e.Data
			if !vectorsClose(aData, eData, prec) {
				t.Errorf("no trans: expected %v but got %v", eData, aData)
			}
		}
	})
}

func TestMatMul(t *testing.T) {
	runWithCreators(t, func(t *testing.T, c anyvec.Creator, prec float64) {
		m2x3 := makeMatrix(c, testMat2x3, 2, 3)
		m3x4 := makeMatrix(c, testMat3x4, 3, 4)
		m2x4 := makeMatrix(c, testMat2x4, 2, 4)
		condAccum := func(m *anydiff.Matrix, accum bool) *anydiff.Matrix {
			if accum {
				d := accumulate(m.Data)
				return &anydiff.Matrix{Data: d, Rows: m.Rows, Cols: m.Cols}
			} else {
				return &anydiff.Matrix{Data: m.Data, Rows: m.Rows, Cols: m.Cols}
			}
		}
		cases := []func(bool) anydiff.Res{
			func(acc bool) anydiff.Res {
				return anydiff.MatMul(false, false, condAccum(m2x3, acc),
					condAccum(m3x4, acc)).Data
			},
			func(acc bool) anydiff.Res {
				return anydiff.MatMul(false, true, condAccum(m3x4, acc),
					condAccum(m2x4, acc)).Data
			},
			func(acc bool) anydiff.Res {
				return anydiff.MatMul(true, false, condAccum(m2x3, acc),
					condAccum(m2x4, acc)).Data
			},
			func(acc bool) anydiff.Res {
				return anydiff.MatMul(true, true, condAccum(m3x4, acc),
					condAccum(m2x3, acc)).Data
			},
		}
		vars := []*anydiff.Var{m2x3.Data.(*anydiff.Var), m3x4.Data.(*anydiff.Var),
			m2x4.Data.(*anydiff.Var)}
		for _, accum := range []bool{false, true} {
			for i, f := range cases {
				t.Run(fmt.Sprintf("Case%v-%d", accum, i), func(t *testing.T) {
					ch := &ResChecker{
						F: func() anydiff.Res {
							return f(accum)
						},
						V: vars,
					}
					ch.FullCheck(t)
				})
			}
		}
	})
}

func TestBatchedMatMul(t *testing.T) {
	runWithCreators(t, func(t *testing.T, c anyvec.Creator, prec float64) {
		m3x2 := makeMatrixBatch(c, testMat3x4, 3, 2, 2)
		m2x2 := makeMatrixBatch(c, testMat2x4, 2, 2, 2)
		vars := []*anydiff.Var{m3x2.Data.(*anydiff.Var), m2x2.Data.(*anydiff.Var)}
		ch := &ResChecker{
			F: func() anydiff.Res {
				return anydiff.BatchedMatMul(false, false, m3x2, m2x2).Data
			},
			V: vars,
		}
		ch.FullCheck(t)
	})
}

func TestSumRows(t *testing.T) {
	runWithCreators(t, func(t *testing.T, c anyvec.Creator, prec float64) {
		m3x4 := makeMatrix(c, testMat3x4, 3, 4)
		ch := &ResChecker{
			F: func() anydiff.Res {
				return anydiff.SumRows(m3x4)
			},
			V: []*anydiff.Var{m3x4.Data.(*anydiff.Var)},
		}
		ch.FullCheck(t)
	})
}

func TestSumCols(t *testing.T) {
	runWithCreators(t, func(t *testing.T, c anyvec.Creator, prec float64) {
		m3x4 := makeMatrix(c, testMat3x4, 3, 4)
		ch := &ResChecker{
			F: func() anydiff.Res {
				return anydiff.SumCols(m3x4)
			},
			V: []*anydiff.Var{m3x4.Data.(*anydiff.Var)},
		}
		ch.FullCheck(t)
	})
}

func TestScaleRowsOut(t *testing.T) {
	runWithCreators(t, func(t *testing.T, c anyvec.Creator, prec float64) {
		m3x4 := makeMatrix(c, testMat3x4, 3, 4)
		scalers := anydiff.NewVar(c.MakeVectorData(c.MakeNumericList([]float64{
			2, -3.5, 1.5,
		})))
		scaled := getComponents(anydiff.ScaleRows(m3x4, scalers).Data.Output())
		expected := []float64{
			0.7125156821328008, 1.5157235964903444, -0.3826236996528368, 3.6904266319483057,
			3.0851091132544637, -3.3609194643252938, 0.0458185048253356, -0.1127736506599714,
			0.4863635876082264, -1.8461093391837804, -0.2338000386485222, 0.7239048071355615,
		}
		if !vectorsClose(scaled, expected, prec) {
			t.Errorf("expected %v but got %v", expected, scaled)
		}
	})
}

func TestScaleRowsProp(t *testing.T) {
	runWithCreators(t, func(t *testing.T, c anyvec.Creator, prec float64) {
		m3x4 := makeMatrix(c, testMat3x4, 3, 4)
		scalers := makeRandomVec(c, 3)
		ch := &ResChecker{
			F: func() anydiff.Res {
				return anydiff.ScaleRows(m3x4, scalers).Data
			},
			V: []*anydiff.Var{m3x4.Data.(*anydiff.Var), scalers},
		}
		ch.FullCheck(t)
	})
}

func TestTranspose(t *testing.T) {
	runWithCreators(t, func(t *testing.T, c anyvec.Creator, prec float64) {
		m2x3 := makeMatrix(c, testMat2x3, 2, 3)
		ch := &ResChecker{
			F: func() anydiff.Res {
				return anydiff.Transpose(m2x3).Data
			},
			V: []*anydiff.Var{m2x3.Data.(*anydiff.Var)},
		}
		ch.FullCheck(t)
	})
}

func makeMatrix(c anyvec.Creator, d []float64, rows, cols int) *anydiff.Matrix {
	return &anydiff.Matrix{
		Data: anydiff.NewVar(c.MakeVectorData(c.MakeNumericList(d))),
		Rows: rows,
		Cols: cols,
	}
}

func makeMatrixBatch(c anyvec.Creator, d []float64, rows, cols, num int) *anydiff.MatrixBatch {
	return &anydiff.MatrixBatch{
		Data: anydiff.NewVar(c.MakeVectorData(c.MakeNumericList(d))),
		Rows: rows,
		Cols: cols,
		Num:  num,
	}
}
