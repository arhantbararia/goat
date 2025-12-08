package node

type Node struct {
	Name            string
	Ip              string
	Cores           int
	Memory          int
	MemoryAllocated int
	DiskAllocated   int
	Role            int
	TaskCount       int
}
