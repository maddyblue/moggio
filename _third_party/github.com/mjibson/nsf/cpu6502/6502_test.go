/*
 * Copyright (c) 2014 Matt Jibson <matt.jibson@gmail.com>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package cpu6502

import (
	"io/ioutil"
	"testing"
)

type Ram []byte

func (r Ram) Read(v uint16) byte     { return r[v] }
func (r Ram) Write(v uint16, b byte) { r[v] = b }

// Download from https://github.com/Klaus2m5/6502_65C02_functional_tests/blob/master/bin_files/6502_functional_test.bin
// GPL, so not included here.
func TestFunctional(t *testing.T) {
	b, err := ioutil.ReadFile("6502_functional_test.bin")
	if err != nil {
		t.Fatal(err)
	}
	r := make(Ram, 0xffff+1)
	copy(r[:], b)
	c := New(r)
	c.L = make([]Log, 20)
	c.PC = 0x0400
	i := 0
	for {
		if c.PC == 0x3399 {
			break
		}
		pc := c.PC
		c.Step()
		if c.PC == pc {
			t.Log(c.StringLog())
			t.Log(c.String())
			t.Log(i, "instructions ran")
			t.Fatalf("repeated PC: 0x%04X", pc)
		} else if c.PC <= 0x1ff {
			t.Log(c.StringLog())
			t.Log(c.String())
			t.Fatalf("low PC: 0x%04X", c.PC)
		}
		i++
	}
}
