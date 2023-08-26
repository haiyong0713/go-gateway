package tool

/*
*
Method: Check the specified string size is over or not
@Param:

	str: the specified string
	limit: max byte for limited

@return:

	isOver:
*/
func IsStringOverSizeLimit(str string, limit int64) (isOver bool) {
	size := int64(len(str) * 4)
	if size > limit {
		isOver = true
	}

	return
}
