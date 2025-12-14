package node

type Node struct {
	Name            string
	Ip              string
	Cores           int
	Memory          int
	Disk            int
	MemoryAllocated int
	DiskAllocated   int
	Role            string
	TaskCount       int
}
