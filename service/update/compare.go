package update

// Return true if at least one of the newstamps of the allstamps at the right place
// In math if there is a i between 0 and perPage -1 that allstamps[(page -1)*perPage + i] = newstamps[i]
// Return false there wise
// Change all the values that don't match that inside allstamps
// TODO: Find a better name for this function
func CompareAndFusion(perPage, page int, newStamps []int64, allstamps *[]int64) bool {
	timestamps := *allstamps
	tailLen := perPage
	if len(newStamps) < perPage {
		// If this batch of timestamps is shorter than a the number of timestamps
		// per page, the size of the slice with all the stamps should adapt to,
		// exactly, the size of all the pages underneath plus the size of this batch
		tailLen = len(newStamps)
		*allstamps = make([]int64, (page-1)*perPage+tailLen)
		copy(*allstamps, timestamps)
	}

	// else, if all the timestamps are fewer than the total amount of pages
	// then it size should grow because it is about to be filled
	if tailLen == perPage && len(timestamps) < page*perPage {
		*allstamps = make([]int64, page*perPage)
		copy(*allstamps, timestamps)
	}

	for i := tailLen - 1; i >= 0; i-- {
		if (*allstamps)[(page-1)*perPage+i] == newStamps[i] {
			return true
		}
		(*allstamps)[(page-1)*perPage+i] = newStamps[i]
	}
	return false
}
