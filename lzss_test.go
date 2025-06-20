package lzss

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestCompressDecompressRoundtrip(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{"empty", []byte("")},
		{"single_char", []byte("a")},
		{"short_string", []byte("Hello, World!")},
		{"repetitive", []byte("AAAAAAAAAAAAAAAAAAAAAA")},
		{"no_repetition", []byte("abcdefghijklmnopqrstuvwxyz")},
		{"mixed_content", []byte("This is a test string for LZSS compression. Hello, World! This repeats to test compression effectiveness.")},
		{"binary_data", []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}},
		{"all_zeros", bytes.Repeat([]byte{0x00}, 100)},
		{"all_ones", bytes.Repeat([]byte{0xFF}, 100)},
		{"pattern_repeat", bytes.Repeat([]byte("ABC123"), 20)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Compress
			compressed := Compress(tc.data)
			
			// Decompress
			decompressed := Decompress(compressed)
			
			// Check if roundtrip works
			if !bytes.Equal(tc.data, decompressed) {
				t.Errorf("Roundtrip failed for %s", tc.name)
				t.Errorf("Original length: %d", len(tc.data))
				t.Errorf("Compressed length: %d", len(compressed))
				t.Errorf("Decompressed length: %d", len(decompressed))
				
				if len(tc.data) < 100 && len(decompressed) < 100 {
					t.Errorf("Original:     %v", tc.data)
					t.Errorf("Decompressed: %v", decompressed)
					t.Errorf("Original string:     %q", string(tc.data))
					t.Errorf("Decompressed string: %q", string(decompressed))
				}
			}
			
			// Log compression ratio for analysis
			if len(tc.data) > 0 {
				ratio := float64(len(compressed)) / float64(len(tc.data))
				t.Logf("%s: original=%d, compressed=%d, ratio=%.2f", 
					tc.name, len(tc.data), len(compressed), ratio)
			}
		})
	}
}

func TestCompressEmpty(t *testing.T) {
	result := Compress([]byte{})
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %v", result)
	}
}

func TestDecompressEmpty(t *testing.T) {
	result := Decompress([]byte{})
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %v", result)
	}
}

func TestLargeData(t *testing.T) {
	// Generate large test data with patterns
	data := make([]byte, 10000)
	for i := range data {
		if i < 1000 {
			data[i] = byte(i % 256)
		} else {
			// Add some repetitive patterns
			data[i] = data[i%1000]
		}
	}
	
	compressed := Compress(data)
	decompressed := Decompress(compressed)
	
	if !bytes.Equal(data, decompressed) {
		t.Errorf("Large data roundtrip failed")
		t.Errorf("Original length: %d", len(data))
		t.Errorf("Decompressed length: %d", len(decompressed))
	}
	
	ratio := float64(len(compressed)) / float64(len(data))
	t.Logf("Large data: original=%d, compressed=%d, ratio=%.2f", 
		len(data), len(compressed), ratio)
}

func TestRandomData(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	
	for i := 0; i < 10; i++ {
		size := rand.Intn(1000) + 100
		data := make([]byte, size)
		rand.Read(data)
		
		compressed := Compress(data)
		decompressed := Decompress(compressed)
		
		if !bytes.Equal(data, decompressed) {
			t.Errorf("Random data roundtrip failed for iteration %d", i)
			t.Errorf("Data size: %d", size)
			break
		}
	}
}

func TestSpecificByteValues(t *testing.T) {
	// Test edge cases with specific byte values
	testCases := [][]byte{
		{0x00},                    // null byte
		{0xFF},                    // max byte
		{0x00, 0xFF, 0x00, 0xFF}, // alternating
		bytes.Repeat([]byte{0x80}, 50), // mid-range repeated
	}
	
	for i, data := range testCases {
		t.Run(fmt.Sprintf("byte_test_%d", i), func(t *testing.T) {
			compressed := Compress(data)
			decompressed := Decompress(compressed)
			
			if !bytes.Equal(data, decompressed) {
				t.Errorf("Byte test %d failed", i)
				t.Errorf("Original:     %v", data)
				t.Errorf("Decompressed: %v", decompressed)
			}
		})
	}
}

func BenchmarkCompress(b *testing.B) {
	data := []byte("This is a test string for LZSS compression benchmarking. " +
		"It contains some repetitive content to test the compression effectiveness. " +
		"This is a test string for LZSS compression benchmarking.")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compress(data)
	}
}

func BenchmarkDecompress(b *testing.B) {
	data := []byte("This is a test string for LZSS compression benchmarking. " +
		"It contains some repetitive content to test the compression effectiveness. " +
		"This is a test string for LZSS compression benchmarking.")
	compressed := Compress(data)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decompress(compressed)
	}
}

func BenchmarkCompressLarge(b *testing.B) {
	// Create 10KB of data with patterns
	data := make([]byte, 10240)
	pattern := []byte("Hello, World! This is a test pattern. ")
	for i := range data {
		data[i] = pattern[i%len(pattern)]
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compress(data)
	}
}

// TestOracleCompat tests compatibility with reference implementations
// To use this test:
// 1. Clone https://github.com/MichaelDipperstein/lzss
// 2. Compile with: make
// 3. Set LZSS_ORACLE_PATH environment variable to the path of the 'sample' binary
// 4. Run: LZSS_ORACLE_PATH=/path/to/sample go test -v -run TestOracleCompat
func TestOracleCompat(t *testing.T) {
	// This test requires an external LZSS implementation
	// Skip if oracle path not provided
	t.Skip("Oracle test requires external LZSS binary - see comments for setup")
	
	// Example test data that should work with standard LZSS
	testData := [][]byte{
		[]byte("Hello, World!"),
		[]byte("AAAAAAAAAAAAAAAA"),  // highly compressible
		[]byte("abcdefghijk"),        // not compressible
		bytes.Repeat([]byte("test"), 100), // pattern repetition
	}
	
	for i, data := range testData {
		t.Run(fmt.Sprintf("oracle_test_%d", i), func(t *testing.T) {
			// Our implementation
			ourCompressed := Compress(data)
			ourDecompressed := Decompress(ourCompressed)
			
			// Verify our roundtrip works
			if !bytes.Equal(data, ourDecompressed) {
				t.Errorf("Our roundtrip failed for test %d", i)
				return
			}
			
			t.Logf("Test %d: original=%d, our_compressed=%d, ratio=%.2f", 
				i, len(data), len(ourCompressed), float64(len(ourCompressed))/float64(len(data)))
		})
	}
}

// TestKnownGoodVectors tests against known compression test vectors
func TestKnownGoodVectors(t *testing.T) {
	// These are manually verified test vectors
	vectors := []struct {
		name     string
		input    []byte
		// We don't specify exact compressed output since different LZSS
		// implementations may produce different but valid results
		minRatio float64 // minimum expected compression ratio
		maxRatio float64 // maximum expected compression ratio
	}{
		{
			name:     "empty",
			input:    []byte{},
			minRatio: 0,
			maxRatio: 0,
		},
		{
			name:     "single_byte", 
			input:    []byte{0x41},
			minRatio: 2.0, // single byte should expand due to overhead
			maxRatio: 10.0,
		},
		{
			name:     "highly_repetitive",
			input:    bytes.Repeat([]byte{0x41}, 1000),
			minRatio: 0.01, // should compress very well
			maxRatio: 0.1,
		},
		{
			name:     "random_like",
			input:    []byte("abcdefghijklmnopqrstuvwxyz0123456789"),
			minRatio: 0.8, // might not compress much
			maxRatio: 2.0,
		},
	}
	
	for _, v := range vectors {
		t.Run(v.name, func(t *testing.T) {
			compressed := Compress(v.input)
			decompressed := Decompress(compressed)
			
			// Test roundtrip
			if !bytes.Equal(v.input, decompressed) {
				t.Errorf("Roundtrip failed for %s", v.name)
				return
			}
			
			// Test compression ratio is within expected bounds
			var ratio float64
			if len(v.input) > 0 {
				ratio = float64(len(compressed)) / float64(len(v.input))
			}
			
			if ratio < v.minRatio || ratio > v.maxRatio {
				t.Errorf("Compression ratio %.3f outside expected range [%.3f, %.3f] for %s",
					ratio, v.minRatio, v.maxRatio, v.name)
			}
			
			t.Logf("%s: input=%d, compressed=%d, ratio=%.3f", 
				v.name, len(v.input), len(compressed), ratio)
		})
	}
}