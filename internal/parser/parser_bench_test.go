package parser

import "testing"

func BenchmarkParseSampleAPI(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Parse(sampleAPI); err != nil {
			b.Fatalf("Parse() error = %v", err)
		}
	}
}
