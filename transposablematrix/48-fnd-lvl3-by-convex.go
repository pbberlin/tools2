package transposablematrix

//
// Find the perfect fit for a given edge x-y-x
func StairyPerfect(ar *Reservoir, fs Fusion) (chosen *Amorph, baseShift Point) {

	var x1, y, x2, directionIdx, maxOffs = fs.xyx[0], fs.xyx[1], fs.xyx[2], fs.dirIdx, fs.maxOffs

	_, chosen = ar.ByStairyEdge(fs, x1, y, x2, maxOffs, VariDirection(directionIdx), grow)
	return chosen, Point{}
}

//
func StairyShrinky(ar *Reservoir, fs Fusion) (chosen *Amorph, baseShift Point) {

	var x1, y, x2, directionIdx, maxOffs = fs.xyx[0], fs.xyx[1], fs.xyx[2], fs.dirIdx, fs.maxOffs
	_, _, _ = x1, y, x2
	_, _ = directionIdx, maxOffs

	return

}

// StraightPerfect tries a perfect fit
func StraightPerfect(ar *Reservoir, fs Fusion) (chosen *Amorph, baseShift Point) {

	var x1, y, x2, directionIdx, maxOffs = fs.xyx[0], fs.xyx[1], fs.xyx[2], fs.dirIdx, fs.maxOffs
	_, _ = directionIdx, maxOffs

	x1, y, x2, baseShift = swapAndAdjustBase(x1, y, x2, baseShift)
	pf("try straight perfect fit%v ", x1)
	_, chosen = exactStraightEdge(ar, x1)
	pf("\n")

	return

}

// StraightShrinky tries to fill
// a *wide* straight concave gap,
// wider than double SmallestDesirableWidth.
func StraightShrinky(ar *Reservoir, fs Fusion) (chosen *Amorph, baseShift Point) {

	// pfTmp := intermedPf(pf)
	// defer func() { pf = pfTmp }()

	var x1, y, x2, directionIdx, maxOffs = fs.xyx[0], fs.xyx[1], fs.xyx[2], fs.dirIdx, fs.maxOffs
	_, _ = directionIdx, maxOffs

	x1, y, x2, baseShift = swapAndAdjustBase(x1, y, x2, baseShift)

	if x1 > wideGapCap*ar.SmallestDesirableWidth {
		x1 = wideGapMin * ar.SmallestDesirableWidth // worsens results
		x1 = ar.SmallestDesirableWidth              // testing yields: best fallback
	}

	if x1 >= wideGapMin*ar.SmallestDesirableWidth {

		pf("gap%v wider than%v => StraightShrinky ", x1, wideGapMin*ar.SmallestDesirableWidth)

		// leave at least SmallestDesirableWidth for further fill
		// stop at SmallestDesirableWidth
		// x-SDW ... SDW
		shrinkStart := x1 - ar.SmallestDesirableWidth
		for i := shrinkStart; i >= ar.SmallestDesirableWidth; i-- {
			pf("srch%v ", i)
			_, chosen = exactStraightEdge(ar, i)
			if chosen != nil {
				// pf(" found %v", chosen)
				break
			}
		}
		pf("\n")
		return chosen, baseShift
	} else {
		pf("gap%v narrower than%v => no StraightShrinky\n", x1, wideGapMin*ar.SmallestDesirableWidth)
		return
	}

}

// ByNumElementsWrap - find amorphs by number of elements,
// search greater or equal (GTE) number than param x1
func ByNumElementsWrap(ar *Reservoir, fs Fusion) (chosen *Amorph, baseShift Point) {

	var x1, y, x2, directionIdx, maxOffs = fs.xyx[0], fs.xyx[1], fs.xyx[2], fs.dirIdx, fs.maxOffs
	_, _ = directionIdx, maxOffs

	x1, y, x2, baseShift = swapAndAdjustBase(x1, y, x2, baseShift)

	if x1 > wideGapCap*ar.SmallestDesirableWidth {
		x1 = ar.SmallestDesirableWidth
	}

	numElementsMin := x1 * ar.SmallestDesirableHeight
	chosen = moreThanXElements(ar, numElementsMin)

	if chosen == nil {
		// rare case: no amorph had more elements
		// than x1*ar.SmallestDesirableHeight
		// => fallback to minimal height of 1
		// then search backward
		for i := 0; i < x1; i++ {
			numElementsMin = x1*minPossibleHeight - i
			chosen = moreThanXElements(ar, numElementsMin)
			if chosen != nil {
				break
			}
		}
	}

	if chosen != nil {
		chosen.Cols = x1
		chosen.Rows, chosen.Slack = OtherSide(chosen.NElements, chosen.Cols)
		numOrig := chosen.NElements

		if chosen.Slack > 0 {
			chosen.Padded = chosen.Slack
			chosen.NElements += chosen.Slack
			chosen.Slack = 0
		}

		if chosen.Padded > 0 {
			pf("srch%v (%v*%v) fnd%v (%v*%v) pad%v \n",
				numElementsMin,
				x1, ar.SmallestDesirableHeight,
				numOrig,
				chosen.Cols, chosen.Rows,
				chosen.Padded)
		}

		if y < 0 {
			baseShift.y += y
		}

		// baseShift.y -= 2
		// baseShift.x++

	} else {
		pf("The next to last matching failed.\n")
	}

	return
}

func swapAndAdjustBase(x1, y, x2 int, baseShift Point) (int, int, int, Point) {

	if y < 0 {
		baseShift.x += x1
		x1, x2 = x2, x1
	}
	return x1, y, x2, baseShift
}
