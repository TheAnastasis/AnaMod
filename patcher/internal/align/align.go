package align

func Up(size, block int) int {
	return (size + block - 1) / block * block
}
