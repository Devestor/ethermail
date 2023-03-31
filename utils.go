package main

import "math/rand"

func randomSample(n, k int) []int {
	if k > n {
		return nil
	}

	nums := make([]int, n)
	for i := range nums {
		nums[i] = i
	}

	rand.Shuffle(len(nums), func(i, j int) {
		nums[i], nums[j] = nums[j], nums[i]
	})

	return nums[:k]
}
