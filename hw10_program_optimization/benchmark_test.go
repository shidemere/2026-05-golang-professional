package hw10programoptimization

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

var benchmarkResult DomainStat

func BenchmarkGetDomainStat(b *testing.B) {
	zr, err := zip.OpenReader("testdata/users.dat.zip")
	if err != nil {
		b.Fatal(err)
	}
	defer zr.Close()

	input, err := zr.File[0].Open()
	if err != nil {
		b.Fatal(err)
	}

	data, err := io.ReadAll(input)
	if err != nil {
		b.Fatal(err)
	}
	if err := input.Close(); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := GetDomainStat(bytes.NewReader(data), "biz")
		if err != nil {
			b.Fatal(err)
		}
		benchmarkResult = result
	}
}
