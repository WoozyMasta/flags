// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

func levenshtein(s string, t string) int {
	sr := []rune(s)
	tr := []rune(t)

	ls := len(sr)
	lt := len(tr)

	if ls == 0 {
		return lt
	}

	if lt == 0 {
		return ls
	}

	dists := make([][]int, ls+1)
	for i := range dists {
		dists[i] = make([]int, lt+1)
		dists[i][0] = i
	}

	for j := 1; j <= lt; j++ {
		dists[0][j] = j
	}

	for i, sc := range sr {
		for j, tc := range tr {
			if sc == tc {
				dists[i+1][j+1] = dists[i][j]
			} else {
				dists[i+1][j+1] = dists[i][j] + 1
				if dists[i+1][j] < dists[i+1][j+1] {
					dists[i+1][j+1] = dists[i+1][j] + 1
				}
				if dists[i][j+1] < dists[i+1][j+1] {
					dists[i+1][j+1] = dists[i][j+1] + 1
				}
			}
		}
	}

	return dists[ls][lt]
}

func closestChoice(cmd string, choices []string) (string, int) {
	if len(choices) == 0 {
		return "", 0
	}

	mincmd := -1
	mindist := -1

	for i, c := range choices {
		l := levenshtein(cmd, c)

		if mincmd < 0 || l < mindist {
			mindist = l
			mincmd = i
		}
	}

	return choices[mincmd], mindist
}
