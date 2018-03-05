package main

import (
	"testing"
)


func BenchmarkTicking(b *testing.B) {
	node := newNode(1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.ticking(1)
	}
}
