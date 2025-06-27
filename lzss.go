// Copyright © 2018 blacktop
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package lzss

import "bytes"

const (
	// n is the size of ring buffer - must be power of 2
	n = 4096
	// f is the upper limit for match_length
	f = 18
	// threshold encode string into position and length if match_length is greater than this
	threshold = 2
	// nil index for root of binary search trees
	nil     = n
	padding = 0x16c
	// Magic and CompressionType for LZSS Apple format
	Magic           = "complzss"
	CompressionType = 0x636f6d70 // "comp"
	Signature       = 0x6c7a7373 // "lzss"
)

// Header represents the LZSS header
type Header struct {
	CompressionType  uint32 // 0x636f6d70 "comp"
	Signature        uint32 // 0x6c7a7373 "lzss"
	CheckSum         uint32 // Likely CRC32
	UncompressedSize uint32
	CompressedSize   uint32
	Padding          [padding]byte
}

// encodeState represents the compression state with binary search trees
type encodeState struct {
	lchild        [n + 1]int
	rchild        [n + 257]int
	parent        [n + 1]int
	textBuf       [n + f - 1]byte
	matchPosition int
	matchLength   int
}

// initState initializes the encoding state, mostly the trees
func (sp *encodeState) initState() {
	// Clear the state
	for i := 0; i < len(sp.lchild); i++ {
		sp.lchild[i] = 0
	}
	for i := 0; i < len(sp.rchild); i++ {
		sp.rchild[i] = 0
	}
	for i := 0; i < len(sp.parent); i++ {
		sp.parent[i] = 0
	}
	for i := 0; i < len(sp.textBuf); i++ {
		sp.textBuf[i] = 0
	}
	sp.matchPosition = 0
	sp.matchLength = 0

	// Initialize buffer with spaces (first N-F positions)
	for i := 0; i < n-f; i++ {
		sp.textBuf[i] = ' '
	}
	// Initialize tree roots to NIL
	for i := n + 1; i <= n+256; i++ {
		sp.rchild[i] = nil
	}
	// Initialize parent pointers to NIL
	for i := 0; i < n; i++ {
		sp.parent[i] = nil
	}
}

// insertNode inserts string of length F into one of the trees and returns the longest match
func (sp *encodeState) insertNode(r int) {
	var i, p, cmp int

	cmp = 1
	key := sp.textBuf[r:]
	p = n + 1 + int(key[0])
	sp.rchild[r] = nil
	sp.lchild[r] = nil
	sp.matchLength = 0

	for {
		if cmp >= 0 {
			if sp.rchild[p] != nil {
				p = sp.rchild[p]
			} else {
				sp.rchild[p] = r
				sp.parent[r] = p
				return
			}
		} else {
			if sp.lchild[p] != nil {
				p = sp.lchild[p]
			} else {
				sp.lchild[p] = r
				sp.parent[r] = p
				return
			}
		}
		for i = 1; i < f; i++ {
			if cmp = int(key[i]) - int(sp.textBuf[p+i]); cmp != 0 {
				break
			}
		}
		if i > sp.matchLength {
			sp.matchPosition = p
			if sp.matchLength = i; sp.matchLength >= f {
				break
			}
		}
	}
	sp.parent[r] = sp.parent[p]
	sp.lchild[r] = sp.lchild[p]
	sp.rchild[r] = sp.rchild[p]
	sp.parent[sp.lchild[p]] = r
	sp.parent[sp.rchild[p]] = r
	if sp.rchild[sp.parent[p]] == p {
		sp.rchild[sp.parent[p]] = r
	} else {
		sp.lchild[sp.parent[p]] = r
	}
	sp.parent[p] = nil
}

// deleteNode deletes node p from tree
func (sp *encodeState) deleteNode(p int) {
	var q int

	if sp.parent[p] == nil {
		return // not in tree
	}
	if sp.rchild[p] == nil {
		q = sp.lchild[p]
	} else if sp.lchild[p] == nil {
		q = sp.rchild[p]
	} else {
		q = sp.lchild[p]
		if sp.rchild[q] != nil {
			for sp.rchild[q] != nil {
				q = sp.rchild[q]
			}
			sp.rchild[sp.parent[q]] = sp.lchild[q]
			sp.parent[sp.lchild[q]] = sp.parent[q]
			sp.lchild[q] = sp.lchild[p]
			sp.parent[sp.lchild[p]] = q
		}
		sp.rchild[q] = sp.rchild[p]
		sp.parent[sp.rchild[p]] = q
	}
	sp.parent[q] = sp.parent[p]
	if sp.rchild[sp.parent[p]] == p {
		sp.rchild[sp.parent[p]] = q
	} else {
		sp.lchild[sp.parent[p]] = q
	}
	sp.parent[p] = nil
}

// Compress compresses data using LZSS algorithm (Apple format)
func Compress(src []byte) []byte {
	if len(src) == 0 {
		return []byte{}
	}

	var dst bytes.Buffer
	sp := &encodeState{}
	var i, c, dataLen, r, s, lastMatchLength, codeBufPtr int
	var codeBuf [17]byte
	var mask byte
	srcPos := 0

	// Initialize trees
	sp.initState()

	// Code buffer setup
	codeBuf[0] = 0
	codeBufPtr = 1
	mask = 1

	// Clear the buffer and set initial positions
	s = 0
	r = n - f

	// Read F bytes into the last F bytes of the buffer
	for dataLen = 0; dataLen < f && srcPos < len(src); dataLen++ {
		sp.textBuf[r+dataLen] = src[srcPos]
		srcPos++
	}
	if dataLen == 0 {
		return []byte{}
	}

	// Insert the F strings, each of which begins with one or more 'space' characters
	for i = 1; i <= f; i++ {
		sp.insertNode(r - i)
	}

	// Finally, insert the whole string just read
	sp.insertNode(r)

	for {
		// match_length may be spuriously long near the end of text
		if sp.matchLength > dataLen {
			sp.matchLength = dataLen
		}
		if sp.matchLength <= threshold {
			sp.matchLength = 1                  // Not long enough match. Send one byte.
			codeBuf[0] |= mask                  // 'send one byte' flag
			codeBuf[codeBufPtr] = sp.textBuf[r] // Send uncoded
			codeBufPtr++
		} else {
			// Send position and length pair. Note match_length > THRESHOLD.
			codeBuf[codeBufPtr] = byte(sp.matchPosition)
			codeBufPtr++
			codeBuf[codeBufPtr] = byte(((sp.matchPosition >> 4) & 0xF0) | (sp.matchLength - (threshold + 1)))
			codeBufPtr++
		}
		if mask <<= 1; mask == 0 { // Shift mask left one bit
			// Send at most 8 units of code together
			for i = 0; i < codeBufPtr; i++ {
				dst.WriteByte(codeBuf[i])
			}
			codeBuf[0] = 0
			codeBufPtr = 1
			mask = 1
		}
		lastMatchLength = sp.matchLength
		for i = 0; i < lastMatchLength && srcPos < len(src); i++ {
			sp.deleteNode(s) // Delete old strings and
			c = int(src[srcPos])
			srcPos++
			sp.textBuf[s] = byte(c) // read new bytes

			// If the position is near the end of buffer, extend the buffer
			// to make string comparison easier.
			if s < f-1 {
				sp.textBuf[s+n] = byte(c)
			}

			// Since this is a ring buffer, increment the position modulo N.
			s = (s + 1) & (n - 1)
			r = (r + 1) & (n - 1)

			// Register the string in text_buf[r..r+F-1]
			sp.insertNode(r)
		}
		for i < lastMatchLength {
			sp.deleteNode(s)

			// After the end of text, no need to read,
			s = (s + 1) & (n - 1)
			r = (r + 1) & (n - 1)
			// but buffer may not be empty.
			dataLen--
			if dataLen > 0 {
				sp.insertNode(r)
			}
			i++
		}
		if dataLen <= 0 {
			break // until length of string to be processed is zero
		}
	}

	if codeBufPtr > 1 { // Send remaining code.
		for i = 0; i < codeBufPtr; i++ {
			dst.WriteByte(codeBuf[i])
		}
	}

	return dst.Bytes()
}

// Decompress decompresses lzss data
func Decompress(src []byte) []byte {
	var i, j, k, r, c int
	var flags uint
	srcPos := 0
	dst := bytes.Buffer{}

	// ring buffer of size n, with extra f-1 bytes to aid string comparison
	textBuf := make([]byte, n+f-1)

	// Initialize ring buffer with spaces (only first N-F positions like C)
	for i = 0; i < n-f; i++ {
		textBuf[i] = ' '
	}

	r = n - f
	flags = 0

	for {
		flags >>= 1
		if (flags & 0x100) == 0 {
			if srcPos >= len(src) {
				break
			}
			c = int(src[srcPos])
			srcPos++
			flags = uint(c | 0xFF00) // uses higher byte cleverly to count eight
		}
		if flags&1 == 1 {
			if srcPos >= len(src) {
				break
			}
			c = int(src[srcPos])
			srcPos++
			dst.WriteByte(byte(c))
			textBuf[r] = byte(c)
			r++
			r &= (n - 1)
		} else {
			if srcPos >= len(src) {
				break
			}
			i = int(src[srcPos])
			srcPos++
			if srcPos >= len(src) {
				break
			}
			j = int(src[srcPos])
			srcPos++
			i |= ((j & 0xF0) << 4)
			j = (j & 0x0F) + threshold
			for k = 0; k <= j; k++ {
				c = int(textBuf[(i+k)&(n-1)])
				dst.WriteByte(byte(c))
				textBuf[r] = byte(c)
				r++
				r &= (n - 1)
			}
		}
	}

	return dst.Bytes()
}
