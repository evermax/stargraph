package update

import "testing"

// TODO: Needs a bit more coverage on other perPage and page number
func TestCompareAndFusion(t *testing.T) {
	tests := []struct {
		perPage        int
		page           int
		oldTstamps     []int64
		newTstamps     []int64
		expected       bool
		expectedStamps []int64
	}{
		{
			perPage:        5,
			page:           3,
			oldTstamps:     []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			newTstamps:     []int64{11, 12, 13, 14, 15},
			expected:       false,
			expectedStamps: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			perPage:        5,
			page:           3,
			oldTstamps:     []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			newTstamps:     []int64{11, 12, 13},
			expected:       false,
			expectedStamps: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13},
		},
		{
			perPage:        5,
			page:           2,
			oldTstamps:     []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			newTstamps:     []int64{6, 7, 8, 10, 11},
			expected:       true,
			expectedStamps: []int64{1, 2, 3, 4, 5, 6, 7, 8, 10, 11},
		},
		{
			perPage:        5,
			page:           2,
			oldTstamps:     []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			newTstamps:     []int64{6, 7, 8},
			expected:       true,
			expectedStamps: []int64{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			perPage:        5,
			page:           2,
			oldTstamps:     []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			newTstamps:     []int64{6, 7, 10},
			expected:       true,
			expectedStamps: []int64{1, 2, 3, 4, 5, 6, 7, 10},
		},
		{
			perPage:        5,
			page:           2,
			oldTstamps:     []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			newTstamps:     []int64{17, 18, 19, 20, 21},
			expected:       false,
			expectedStamps: []int64{1, 2, 3, 4, 5, 17, 18, 19, 20, 21},
		},
		{
			perPage:        5,
			page:           2,
			oldTstamps:     []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 24, 25, 26},
			newTstamps:     []int64{17, 18, 19, 20, 21},
			expected:       false,
			expectedStamps: []int64{1, 2, 3, 4, 5, 17, 18, 19, 20, 21, 24, 25, 26},
		},
		{
			perPage:        5,
			page:           2,
			oldTstamps:     []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 24, 25, 26},
			newTstamps:     []int64{6, 18, 19, 20, 21},
			expected:       true,
			expectedStamps: []int64{1, 2, 3, 4, 5, 6, 18, 19, 20, 21, 24, 25, 26},
		},
		{
			perPage:        5,
			page:           2,
			oldTstamps:     []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 24, 25, 26, 27, 29},
			newTstamps:     []int64{17, 18, 19},
			expected:       false,
			expectedStamps: []int64{1, 2, 3, 4, 5, 17, 18, 19},
		},
		{
			perPage:        5,
			page:           2,
			oldTstamps:     []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 24, 25, 26, 27, 29},
			newTstamps:     []int64{6, 18, 19},
			expected:       true,
			expectedStamps: []int64{1, 2, 3, 4, 5, 6, 18, 19},
		},
		{
			perPage:        5,
			page:           1,
			oldTstamps:     []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 24, 25, 26, 27, 29},
			newTstamps:     []int64{6, 18, 19},
			expected:       false,
			expectedStamps: []int64{6, 18, 19},
		},
		{
			perPage:        5,
			page:           1,
			oldTstamps:     []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 24, 25, 26, 27, 29},
			newTstamps:     []int64{1, 18, 19},
			expected:       true,
			expectedStamps: []int64{1, 18, 19},
		},
	}

	for i, test := range tests {
		if CompareAndFusion(test.perPage, test.page, test.newTstamps, &test.oldTstamps) != test.expected {
			t.Fatalf("Test %d: should be %v\n", i, test.expected)
		}
		if len(test.expectedStamps) != len(test.oldTstamps) {
			t.Fatalf("Test %d: Stamps don't have the same length: should be %v, got %v\n", i, test.expectedStamps, test.oldTstamps)
		}
		for i, stamp := range test.oldTstamps {
			if stamp != test.expectedStamps[i] {
				t.Fatalf("Test %d: Expected slice %v, got %v\n", i, test.expectedStamps, test.oldTstamps)
			}
		}
	}
}
