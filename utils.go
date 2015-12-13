package main

func calcFFTBins(fftSize int) (bins []float64) {
	bins = make([]float64, fftSize/2)
	step := float64(samplesPerSecond) / float64(fftSize)
	for idx := range bins {
		bins[idx] = step * float64(idx)
	}
	return bins
}

func normalizeInPlace(input []float64) {
	var total float64
	for _, val := range input {
		total += val
	}
	for idx := range input {
		input[idx] /= total
	}
}
