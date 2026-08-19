package sorted

func SelectSorted(arr []int) []int {
	n := len(arr)

	for i := 0; i < n; i++ {

		for j := i + 1; j < n; j++ {
			if arr[j] < arr[i] {
				arr[i], arr[j] = arr[j], arr[i]

			}
		}

	}

	return arr
}
