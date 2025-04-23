package wildcard

const STAR = byte('*')

func Matches(pattern []byte, s string) bool {
	piter := pattern
	siter := []byte(s)

	pi, si := 0, 0

	next := func(iter []byte, idx *int) (byte, bool) {
		if *idx < len(iter) {
			ch := iter[*idx]
			*idx++
			return ch, true
		}

		return 0, false
	}

	pc, pOk := next(piter, &pi)
	sc, sOk := next(siter, &si)

	for pOk && sOk {
		if pc == STAR {
			pc, pOk = next(piter, &pi)

			if pOk {
				if pc == STAR {
					continue
				}

				for sOk && sc != pc {
					sc, sOk = next(siter, &si)
				}

				if !sOk {
					return false
				}
			} else {
				return true
			}
		} else if pc != sc {
			return false
		}

		pc, pOk = next(piter, &pi)
		sc, sOk = next(siter, &si)
	}

	if pOk || sOk {
		return false
	}

	return true
}
